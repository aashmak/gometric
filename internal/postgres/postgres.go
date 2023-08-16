package postgres

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	DB *pgxpool.Pool
}

func NewPostgresDB(db *pgxpool.Pool) *Postgres {
	return &Postgres{
		DB: db,
	}
}

func (p *Postgres) InitDB() error {
	queryStr := `CREATE TABLE IF NOT EXISTS metrics (
		name text unique not null,
		type text not null,
		delta integer,
		value double precision
	  );`

	_, err := p.DB.Exec(context.Background(), queryStr)

	return err
}

func (p *Postgres) Clear() error {
	_, err := p.DB.Exec(context.Background(), `TRUNCATE metrics;`)

	return err
}

func (p *Postgres) Set(k string, v interface{}) error {
	if v == nil {
		return fmt.Errorf("invalid value")
	}

	var vtype, column, queryStr string

	switch v.(type) {
	case float64:
		vtype = "float64"
		column = "value"
	case int:
		vtype = "int"
		column = "delta"
	case int64:
		vtype = "int64"
		column = "delta"
	default:
		return fmt.Errorf("invalid type value")
	}

	queryStr = `INSERT INTO metrics (name, type, ` + column + `) VALUES ($1, $2, $3) 
		ON CONFLICT (name) DO UPDATE SET type=$2, ` + column + `=$3`

	_, err := p.DB.Exec(context.Background(), queryStr, k, vtype, v)

	return err
}

func (p *Postgres) MSet(data map[string]interface{}) error {
	//check valid data
	for k, v := range data {
		if k == "" || v == nil {
			return fmt.Errorf("key or value no must be empty")
		}
	}

	ctx := context.Background()

	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for k, v := range data {
		var vtype, column, queryStr string

		switch v.(type) {
		case float64:
			vtype = "float64"
			column = "value"
		case int:
			vtype = "int"
			column = "delta"
		case int64:
			vtype = "int64"
			column = "delta"
		default:
			return fmt.Errorf("invalid type value")
		}

		queryStr = `INSERT INTO metrics (name, type, ` + column + `) VALUES ($1, $2, $3) 
				ON CONFLICT (name) DO UPDATE SET type=$2, ` + column + `=$3`

		_, err := p.DB.Exec(context.Background(), queryStr, k, vtype, v)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (p *Postgres) Get(k string) (interface{}, error) {
	var vtype string
	var delta *int64
	var value *float64

	err := p.DB.QueryRow(context.Background(), `SELECT type, delta, value FROM metrics WHERE name=$1 LIMIT 1;`, k).Scan(&vtype, &delta, &value)
	if err != nil {
		return nil, fmt.Errorf("metric %s not found", k)
	}

	switch vtype {
	case "float64":
		return float64(*value), nil
	case "int":
		return int(*delta), nil
	case "int64":
		return int64(*delta), nil
	default:
		return value, nil
	}
}

func (p *Postgres) List() []string {
	var s []string

	rows, err := p.DB.Query(context.Background(), `SELECT name FROM metrics;`)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return nil
	}

	for rows.Next() {
		var l string
		err := rows.Scan(&l)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			return nil
		}
		s = append(s, l)
	}

	sort.Strings(s)
	return s
}

func (p *Postgres) Close() error {
	p.DB.Close()

	return nil
}

func (p *Postgres) Ping() error {
	return p.DB.Ping(context.Background())
}
