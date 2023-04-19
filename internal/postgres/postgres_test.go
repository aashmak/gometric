package postgres

import (
	"reflect"
	"testing"
)

func TestPostgresDB(t *testing.T) {
	var err error
	dsn := "postgresql://postgres:postgres@postgres:5432/praktikum"
	pg := NewPostgresDB(dsn)
	pg.Open()
	pg.InitDB()
	pg.Clear()
	defer pg.DB.Close()

	err = pg.Set("testFloat1", float64(3.1414))
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = pg.Set("testInt1", int(10))
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = pg.Set("testInt2", int64(10))
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	v, err := pg.Get("testFloat1")
	if err != nil || v != float64(3.1414) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt1")
	if err != nil || v != int(10) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt2")
	if err != nil || v != int64(10) {
		t.Errorf("Error: %s", err)
	}

	a := []string{"testFloat1", "testInt1", "testInt2"}
	b := pg.List()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Error: value is incorrect")
	}
}

func TestPostgresDB_Tx(t *testing.T) {
	var err error
	dsn := "postgresql://postgres:postgres@postgres:5432/praktikum"
	pg := NewPostgresDB(dsn)
	pg.Open()
	pg.InitDB()
	pg.Clear()
	defer pg.DB.Close()

	mdata := make(map[string]interface{})
	mdata["testFloat11"] = float64(1.11)
	mdata[""] = int(11)
	mdata["testInt111"] = int64(11)

	err = pg.MSet(mdata)
	if err == nil {
		t.Errorf("Error: key or value no must be empty")
	}

	mdata1 := make(map[string]interface{})
	mdata1["testFloat22"] = float64(2.22)
	mdata1["testInt22"] = int(22)
	mdata1["testInt222"] = nil

	err = pg.MSet(mdata1)
	if err == nil {
		t.Errorf("Error: key or value no must be empty")
	}

	mdata2 := make(map[string]interface{})
	mdata2["testFloat1"] = float64(3.1414)
	mdata2["testInt1"] = int(10)
	mdata2["testInt2"] = int64(10)

	err = pg.MSet(mdata2)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	v, err := pg.Get("testFloat1")
	if err != nil || v != float64(3.1414) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt1")
	if err != nil || v != int(10) {
		t.Errorf("Error: %s", err)
	}

	v, err = pg.Get("testInt2")
	if err != nil || v != int64(10) {
		t.Errorf("Error: %s", err)
	}

	a := []string{"testFloat1", "testInt1", "testInt2"}
	b := pg.List()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Error: value is incorrect")
	}
}
