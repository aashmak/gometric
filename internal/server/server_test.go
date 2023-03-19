package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, _ := http.NewRequest(method, ts.URL+path, nil)

	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestVariableType(t *testing.T) {
	var c counter
	if !counterType(c) {
		t.Errorf("Error: Variable is counter type.")
	}

	var g gauge
	if !gaugeType(g) {
		t.Errorf("Error: Variable is gauge type.")
	}
}

func TestChiRouter(t *testing.T) {
	s := NewServer()
	s.chiRouter.Route("/", func(router chi.Router) {
		s.chiRouter.Get("/", s.listHandler)
		s.chiRouter.Post("/", s.defaultHandler)
		s.chiRouter.Get("/value/{metricType}/{metricName}", s.GetValueHandler)
		s.chiRouter.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateHandler)
	})

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "GET", "/")
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	statusCode, body = testRequest(t, ts, "POST", "/")
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/PollCount/1")
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/PollCount/2")
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	statusCode, body = testRequest(t, ts, "GET", "/value/counter/PollCount")
	if statusCode != http.StatusOK || body != "3" {
		t.Errorf("Error")
	}
}

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
