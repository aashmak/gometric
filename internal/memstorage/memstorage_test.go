package memstorage

import (
	"reflect"
	"testing"
)

func TestStorage1(t *testing.T) {
	memStor := NewMemStorage()

	_, err := memStor.Get("abc")
	if err == nil {
		t.Errorf("Error: key is not exist")
	}
}

func TestStorage2(t *testing.T) {
	memStor := NewMemStorage()

	err := memStor.Set("abc", int(1))
	if err != nil {
		t.Errorf("Error: key is not saved")
	}

	if v, _ := memStor.Get("abc"); v != int(1) {
		t.Errorf("Error: value is incorrect")
	}

	memStor.Set("abc", int(2))
	if v, _ := memStor.Get("abc"); v != int(2) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestStorage3(t *testing.T) {
	memStor := NewMemStorage()

	err := memStor.Set("abc", float64(3.14))
	if err != nil {
		t.Errorf("Error: key is not saved")
	}

	if v, _ := memStor.Get("abc"); v != float64(3.14) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestStorage4(t *testing.T) {
	memStor := NewMemStorage()

	err := memStor.Set("abc", "abc")
	if err != nil {
		t.Errorf("Error: key is not saved")
	}

	if v, _ := memStor.Get("abc"); v != "abc" {
		t.Errorf("Error: value is incorrect")
	}
}

func TestStorage5(t *testing.T) {
	memStor := NewMemStorage()

	memStor.Set("xyz", "abc")
	memStor.Set("abc", "abc")
	memStor.Set("def", int(1))

	a := []string{"abc", "def", "xyz"}
	b := memStor.List()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Error: value is incorrect")
	}
}
