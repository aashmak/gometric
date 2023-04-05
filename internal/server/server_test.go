package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte) (int, string) {
	req, _ := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func testRequestGzip(t *testing.T, ts *httptest.Server, method, path string, body []byte) (int, string) {
	var b bytes.Buffer

	g := gzip.NewWriter(&b)
	_, err := g.Write(body)
	if err != nil {
		log.Fatalf("failed init compress writer: %v", err)
	}
	g.Close()

	req, _ := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(b.Bytes()))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, _ := http.DefaultClient.Do(req)

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("failed init compress reader: %v", err)
		}
		defer reader.Close()

		resp.Body = io.NopCloser(reader)
	}

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

// test without gzip
func TestChiRouter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	s := NewServer(ctx, cfg)
	s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))
	s.chiRouter.Use(unzipBodyHandler)
	s.chiRouter.Get("/", s.listHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Post("/value/", s.GetValueHandler)
	s.chiRouter.Post("/update/", s.UpdateHandler)

	var reqBody []byte
	var body string
	var statusCode int

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()

	statusCode, _ = testRequest(t, ts, "GET", "/", nil)
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	statusCode, _ = testRequest(t, ts, "POST", "/", nil)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	// test update gauge
	reqBody = []byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":100}")
	statusCode, _ = testRequest(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	// test get gauge
	reqBody = []byte("{\"id\":\"Alloc\",\"type\":\"gauge\"}")
	statusCode, body = testRequest(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusOK || body != "100" {
		t.Errorf("Error")
	}

	// test insupported type
	reqBody = []byte("{\"id\":\"Alloc\",\"type\":\"counter\"}")
	statusCode, _ = testRequest(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusNotFound {
		t.Errorf("Error")
	}

	// test update counter
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":1}")
	statusCode, _ = testRequest(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	// test get counter
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter\"}")
	statusCode, body = testRequest(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusOK || body != "1" {
		t.Errorf("Error")
	}

	// test insupported type
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"gauge\"}")
	statusCode, _ = testRequest(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusNotFound {
		t.Errorf("Error")
	}

	// === test invalid counter ===
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter\"}")
	statusCode, _ = testRequest(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter111\"}")
	statusCode, _ = testRequest(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	reqBody = []byte("{}")
	statusCode, _ = testRequest(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	reqBody = []byte("{\"id\":\"\",\"type\":\"counter\",\"delta\":1}")
	statusCode, _ = testRequest(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}
}

// test with gzip
func TestChiRouterGzip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	s := NewServer(ctx, cfg)
	s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))
	s.chiRouter.Use(unzipBodyHandler)
	s.chiRouter.Get("/", s.listHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Post("/value/", s.GetValueHandler)
	s.chiRouter.Post("/update/", s.UpdateHandler)

	var reqBody []byte
	var body string
	var statusCode int

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()

	statusCode, _ = testRequestGzip(t, ts, "GET", "/", nil)
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	statusCode, _ = testRequestGzip(t, ts, "POST", "/", nil)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	// test update gauge
	reqBody = []byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":100}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	// test get gauge
	reqBody = []byte("{\"id\":\"Alloc\",\"type\":\"gauge\"}")
	statusCode, body = testRequestGzip(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusOK || body != "100" {
		t.Errorf("Error")
	}

	// test insupported type
	reqBody = []byte("{\"id\":\"Alloc\",\"type\":\"counter\"}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusNotFound {
		t.Errorf("Error")
	}

	// test update counter
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":1}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusOK {
		t.Errorf("Error")
	}

	// test get counter
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter\"}")
	statusCode, body = testRequestGzip(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusOK || body != "1" {
		t.Errorf("Error")
	}

	// test insupported type
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"gauge\"}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/value/", reqBody)
	if statusCode != http.StatusNotFound {
		t.Errorf("Error")
	}

	// === test invalid counter ===
	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter\"}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	reqBody = []byte("{\"id\":\"PollCount\",\"type\":\"counter111\"}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	reqBody = []byte("{}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}

	reqBody = []byte("{\"id\":\"\",\"type\":\"counter\",\"delta\":1}")
	statusCode, _ = testRequestGzip(t, ts, "POST", "/update/", reqBody)
	if statusCode != http.StatusForbidden {
		t.Errorf("Error")
	}
}
