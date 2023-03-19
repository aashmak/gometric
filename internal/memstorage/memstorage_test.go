package memstorage

import (
	"reflect"
	"testing"
)

func TestStorage1(t *testing.T) {
	mem_stor := NewMemStorage()

	_, err := mem_stor.Get("abc")
	if err == nil {
		t.Errorf("Error: key is not exist")
	}
}

func TestStorage2(t *testing.T) {
	mem_stor := NewMemStorage()

	err := mem_stor.Set("abc", int(1))
	if err != nil {
		t.Errorf("Error: key is not saved")
	}

	if v, _ := mem_stor.Get("abc"); v != int(1) {
		t.Errorf("Error: value is incorrect")
	}

	mem_stor.Set("abc", int(2))
	if v, _ := mem_stor.Get("abc"); v != int(2) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestStorage3(t *testing.T) {
	mem_stor := NewMemStorage()

	err := mem_stor.Set("abc", float64(3.14))
	if err != nil {
		t.Errorf("Error: key is not saved")
	}

	if v, _ := mem_stor.Get("abc"); v != float64(3.14) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestStorage4(t *testing.T) {
	mem_stor := NewMemStorage()

	err := mem_stor.Set("abc", "abc")
	if err != nil {
		t.Errorf("Error: key is not saved")
	}

	if v, _ := mem_stor.Get("abc"); v != "abc" {
		t.Errorf("Error: value is incorrect")
	}
}

func TestStorage5(t *testing.T) {
	mem_stor := NewMemStorage()

	mem_stor.Set("xyz", "abc")
	mem_stor.Set("abc", "abc")
	mem_stor.Set("def", int(1))

	a := []string{"abc", "def", "xyz"}
	b := mem_stor.List()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Error: value is incorrect")
	}
}
