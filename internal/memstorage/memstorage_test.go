package memstorage

import (
	"bufio"
	"os"
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

func TestStorage6(t *testing.T) {
	memStor := NewMemStorage()

	err := memStor.Set("abc", nil)
	if err == nil {
		t.Errorf("Error: %s", err)
	}
}

func TestSetStoreFile(t *testing.T) {
	memStor := NewMemStorage()
	err := memStor.SetStoreFile("")
	if err == nil {
		t.Errorf("Error: filename must not be empty")
	}
}

func TestSetSyncMode(t *testing.T) {
	memStor := NewMemStorage()

	memStor.SetSyncMode(true)
	if !memStor.SyncMode {
		t.Errorf("Error: syn mode must be true")
	}

	memStor.SetSyncMode(false)
	if memStor.SyncMode {
		t.Errorf("Error: syn mode must be false")
	}
}

func TestSaveLoadDump(t *testing.T) {
	storeFile := "/tmp/test_storeFile.json"
	memStor := NewMemStorage()
	memStor.SetStoreFile(storeFile)

	memStor.Set("a", int(1))
	memStor.Set("b", float64(3.14))
	memStor.Set("c", "foo")

	err := memStor.SaveDump()
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	file, err := os.OpenFile(storeFile, os.O_RDONLY, 0777)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	defer file.Close()
	defer os.Remove(storeFile)

	reader := bufio.NewReader(file)
	str, _ := reader.ReadBytes('\n')
	strTest := []byte(`{"a":1,"b":3.14,"c":"foo"}`)
	if !reflect.DeepEqual(str, strTest) {
		t.Errorf("Error: %s", err)
	}

	//test load dump
	dataTmp, err := memStor.LoadDump()
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if dataTmp["a"] != float64(1) || dataTmp["b"] != float64(3.14) || dataTmp["c"] != "foo" {
		t.Errorf("Error: invalid data")
	}
}
