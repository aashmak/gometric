package storage

import "gometric/internal/memstorage"

type Storage interface {
	Set(k string, v interface{}) error
	Get(k string) (interface{}, error)
	List() []string
}

func New() *memstorage.MemStorage {
	return memstorage.NewMemStorage()
}
