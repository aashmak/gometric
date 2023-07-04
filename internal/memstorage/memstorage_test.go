package memstorage

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestOpenClose(t *testing.T) {
	memStor := NewMemStorage()
	err := memStor.Open()
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = memStor.Close()
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}

func TestStorage1(t *testing.T) {
	memStor := NewMemStorage()
	memStor.Open()
	defer memStor.Close()

	_, err := memStor.Get("abc")
	if err == nil {
		t.Errorf("Error: key is not exist")
	}
}

func TestStorage2(t *testing.T) {
	memStor := NewMemStorage()
	memStor.Open()
	defer memStor.Close()

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
	memStor.Open()
	defer memStor.Close()

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
	memStor.Open()
	defer memStor.Close()

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
	memStor.Open()
	defer memStor.Close()

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
	memStor.Open()
	defer memStor.Close()

	err := memStor.Set("abc", nil)
	if err == nil {
		t.Errorf("Error: %s", err)
	}
}

func TestStorage7(t *testing.T) {
	memStor := NewMemStorage()
	memStor.Open()
	defer memStor.Close()

	mData := make(map[string]interface{})
	mData["abc"] = "abc"
	mData["def"] = int(1)
	mData["xyz"] = float64(3.14)

	memStor.MSet(mData)

	if v, _ := memStor.Get("abc"); v != "abc" {
		t.Errorf("Error: value is incorrect")
	}

	if v, _ := memStor.Get("def"); v != int(1) {
		t.Errorf("Error: value is incorrect")
	}

	if v, _ := memStor.Get("xyz"); v != float64(3.14) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestSaveLoadDump(t *testing.T) {
	storeFile := "/tmp/test_storeFile.json"
	memStor := NewMemStorage()
	memStor.StoreFile = storeFile
	memStor.Open()
	defer memStor.Close()

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

func BenchmarkStorage1(b *testing.B) {
	memStor := NewMemStorage()
	memStor.Open()
	defer memStor.Close()

	for i := 0; i < b.N; i++ {
		memStor.Set("abc", int(1))
		memStor.Get("abc")
	}
}

func BenchmarkSaveDump(b *testing.B) {
	memStor := NewMemStorage()
	memStor.Open()
	defer memStor.Close()

	memStor.Set("a", int(1))
	memStor.Set("b", float64(3.14))
	memStor.Set("c", "foo")

	for i := 0; i < b.N; i++ {
		memStor.SaveDump()
	}
}

func ExampleMemStorage_Set() {
	memStor := NewMemStorage()

	memStor.Set("abc", int(1))
	v, _ := memStor.Get("abc")

	fmt.Printf("%d", v)

	// Output:
	// 1
}

func ExampleMemStorage_MSet() {
	memStor := NewMemStorage()
	memStor.Open()
	defer memStor.Close()

	mData := make(map[string]interface{})
	mData["abc"] = "abc"
	mData["def"] = int(1)
	mData["xyz"] = float64(3.14)

	memStor.MSet(mData)

	if v, err := memStor.Get("abc"); err == nil {
		fmt.Printf("%s\n", v)
	}

	if v, err := memStor.Get("def"); err == nil {
		fmt.Printf("%d\n", v)
	}

	if v, err := memStor.Get("xyz"); err == nil {
		fmt.Printf("%.2f", v)
	}

	// Output:
	// abc
	// 1
	// 3.14
}
