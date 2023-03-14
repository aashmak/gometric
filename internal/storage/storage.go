package storage

import (
	"errors"
	"fmt"
	"sync"
)

type gauge float64
type counter int64

type MemStorage struct {
	Mutex   sync.Mutex
	Metrics map[string]interface{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]interface{}),
	}
}

func (m *MemStorage) Set(k string, v interface{}) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.Metrics[k] = v

	return nil
}

func (m *MemStorage) Get(k string) (interface{}, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if v, ok := m.Metrics[k]; !ok {
		return nil, errors.New(fmt.Sprintf("Metric %s not found", k))
	} else if v == nil {
		return 0, nil
	}

	return m.Metrics[k], nil
}
