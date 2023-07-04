package postgres

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresDB(t *testing.T) {
	var db *pgxpool.Pool
	var err error

	dsn := "postgresql://postgres:postgres@postgres:5432/praktikum"
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	pg := NewPostgresDB(db)
	pg.InitDB()
	pg.Clear()
	defer pg.DB.Close()

	err = pg.Set("testFloat1", float64(3.1414))
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = pg.Set("testInt1", int(10))
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = pg.Set("testInt2", int64(10))
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	v, err := pg.Get("testFloat1")
	if err != nil || v != float64(3.1414) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt1")
	if err != nil || v != int(10) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt2")
	if err != nil || v != int64(10) {
		t.Errorf("Error: %s", err)
	}

	a := []string{"testFloat1", "testInt1", "testInt2"}
	b := pg.List()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestPostgresDB_Tx(t *testing.T) {
	var db *pgxpool.Pool
	var err error

	dsn := "postgresql://postgres:postgres@postgres:5432/praktikum"
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	pg := NewPostgresDB(db)
	pg.InitDB()
	pg.Clear()
	defer pg.DB.Close()

	mdata := make(map[string]interface{})
	mdata["testFloat11"] = float64(1.11)
	mdata[""] = int(11)
	mdata["testInt111"] = int64(11)

	err = pg.MSet(mdata)
	if err == nil {
		t.Errorf("Error: key or value no must be empty")
	}

	mdata1 := make(map[string]interface{})
	mdata1["testFloat22"] = float64(2.22)
	mdata1["testInt22"] = int(22)
	mdata1["testInt222"] = nil

	err = pg.MSet(mdata1)
	if err == nil {
		t.Errorf("Error: key or value no must be empty")
	}

	mdata2 := make(map[string]interface{})
	mdata2["testFloat1"] = float64(3.1414)
	mdata2["testInt1"] = int(10)
	mdata2["testInt2"] = int64(10)

	err = pg.MSet(mdata2)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	v, err := pg.Get("testFloat1")
	if err != nil || v != float64(3.1414) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt1")
	if err != nil || v != int(10) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt2")
	if err != nil || v != int64(10) {
		t.Errorf("Error: %s", err)
	}

	a := []string{"testFloat1", "testInt1", "testInt2"}
	b := pg.List()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Error: value is incorrect")
	}
}

func Example() {
	var db *pgxpool.Pool
	var err error

	dsn := "postgresql://postgres:postgres@postgres:5432/praktikum"
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("parse config error")
	}

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("create new pool error")
	}

	pg := NewPostgresDB(db)
	pg.InitDB()
	pg.Clear()
	defer pg.DB.Close()

	mdata := make(map[string]interface{})
	mdata["abc"] = "abc"
	mdata["def"] = int(1)
	mdata["xyz"] = float64(3.14)

	pg.MSet(mdata)
	pg.Set("abcabc", int(2))

	if v, err := pg.Get("abc"); err == nil {
		fmt.Printf("%s\n", v)
	}

	if v, err := pg.Get("def"); err == nil {
		fmt.Printf("%d\n", v)
	}

	if v, err := pg.Get("xyz"); err == nil {
		fmt.Printf("%.2f\n", v)
	}

	if v, err := pg.Get("abcabc"); err == nil {
		fmt.Printf("%d\n", v)
	}

	// Output:
	// abc
	// 1
	// 3.14
	// 2
}
