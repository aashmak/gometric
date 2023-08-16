package storage

import (
	"gometric/internal/memstorage"
	"gometric/internal/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Close() error
	Set(k string, v interface{}) error
	MSet(data map[string]interface{}) error
	Get(k string) (interface{}, error)
	List() []string
}

func NewMemStorage(storeFile string, syncMode bool) (*memstorage.MemStorage, error) {
	m := memstorage.NewMemStorage()
	m.StoreFile = storeFile
	m.SyncMode = syncMode
	err := m.Open()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewPostgresDB(db *pgxpool.Pool) *postgres.Postgres {
	return postgres.NewPostgresDB(db)
}
