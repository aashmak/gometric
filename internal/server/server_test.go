package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMethod(t *testing.T) {
	serv := NewServer()

	wr := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("Method Get not allowed")
	}

	req = httptest.NewRequest(http.MethodGet, "/update/", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("Method Get not allowed")
	}

	req = httptest.NewRequest(http.MethodGet, "/update/gauge/Alloc/1.00", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("Method Get not allowed")
	}
}

func TestPostMethod(t *testing.T) {
	serv := NewServer()

	wr := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/1", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/update/", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/update/gauge", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/1.00", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code == http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}
}

func TestUpdate(t *testing.T) {
	serv := NewServer()

	wr := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/1.00", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code != http.StatusOK {
		t.Errorf("status code != 200")
	}

	req = httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/2.00", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code != http.StatusOK {
		t.Errorf("status code != 200")
	}

	req = httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/1", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code != http.StatusOK {
		t.Errorf("status code != 200")
	}

	req = httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/2", nil)
	serv.UpdateHandler(wr, req)
	if wr.Code != http.StatusOK {
		t.Errorf("status code != 200")
	}
}
