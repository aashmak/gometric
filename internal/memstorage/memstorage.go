package memstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

type MemStorage struct {
	Mutex     sync.Mutex
	StoreFile string
	File      *os.File
	SyncMode  bool
	Metrics   map[string]interface{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		StoreFile: "/tmp/memstorage.json",
		SyncMode:  false,
		Metrics:   make(map[string]interface{}),
	}
}

func (m *MemStorage) Open() error {
	var err error

	m.File, err = os.OpenFile(m.StoreFile, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	return nil
}

func (m *MemStorage) Close() error {
	return m.File.Close()
}

func (m *MemStorage) Set(k string, v interface{}) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if v == nil {
		return fmt.Errorf("invalid value")
	}

	m.Metrics[k] = v

	if m.SyncMode {
		err := m.SaveDump()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MemStorage) MSet(data map[string]interface{}) error {
	//check valid data
	for k, v := range data {
		if k == "" || v == nil {
			return fmt.Errorf("key or value no must be empty")
		}
	}

	m.Mutex.Lock()
	for k, v := range data {
		m.Metrics[k] = v
	}
	m.Mutex.Unlock()

	if m.SyncMode {
		err := m.SaveDump()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MemStorage) Get(k string) (interface{}, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if v, ok := m.Metrics[k]; !ok {
		return nil, fmt.Errorf("metric %s not found", k)
	} else if v == nil {
		return 0, nil
	}

	return m.Metrics[k], nil
}

func (m *MemStorage) List() []string {
	var s []string

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for k := range m.Metrics {
		s = append(s, k)
	}

	sort.Strings(s)
	return s
}

func (m *MemStorage) SaveDump() error {

	db, err := json.Marshal(m.Metrics)
	if err != nil {
		return err
	}

	m.File.Truncate(0)
	m.File.Seek(0, 0)
	m.File.Write(db)
	m.File.Sync()

	return nil
}

func (m *MemStorage) LoadDump() (map[string]interface{}, error) {
	dataTmp := make(map[string]interface{})

	file, err := os.OpenFile(m.StoreFile, os.O_RDONLY, 0777)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	data, _ := reader.ReadBytes('\n')

	err = json.Unmarshal(data, &dataTmp)
	if err != nil {
		return nil, err
	}

	return dataTmp, nil
}
