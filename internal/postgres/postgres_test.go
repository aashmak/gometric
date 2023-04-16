package postgres

import (
	"reflect"
	"testing"
)

func TestPostgresDB(t *testing.T) {
	var err error
	dsn := "postgresql://gometric:123456@172.17.0.2:5432/gometric"
	pg := NewPostgresDB(dsn)
	pg.Open()
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
