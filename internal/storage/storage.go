package storage

import (
	"gometric/internal/memstorage"
	"gometric/internal/postgres"
)

type Storage interface {
	Open() error
	Close() error
	Set(k string, v interface{}) error
	Get(k string) (interface{}, error)
	List() []string
}

func NewMemStorage(storeFile string, syncMode bool) *memstorage.MemStorage {
	m := memstorage.NewMemStorage()
	m.SetStoreFile(storeFile)
	m.SetSyncMode(syncMode)

	return m
}

func NewPostgresDB(dsn string) *postgres.Postgres {
	return postgres.NewPostgresDB(dsn)
}
