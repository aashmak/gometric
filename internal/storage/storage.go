package storage

import "gometric/internal/memstorage"

type Storage interface {
	Set(k string, v interface{}) error
	Get(k string) (interface{}, error)
	List() []string
	LoadDump() (map[string]interface{}, error)
	SaveDump() error
	Open() error
	Close() error
	SetStoreFile(filename string) error
	SetSyncMode(mode bool)
}

func New() *memstorage.MemStorage {
	return memstorage.NewMemStorage()
}
