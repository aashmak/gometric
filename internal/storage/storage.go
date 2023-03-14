package storage

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

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

func (m *MemStorage) List() []string {
	var s []string

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for k, _ := range m.Metrics {
		s = append(s, k)
	}

	sort.Strings(s)
	return s
}
