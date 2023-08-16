package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"gometric/internal/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func httpRequestRealIP(ts *httptest.Server, method, path string, body []byte, realIP string) (int, string) {
	req, _ := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	if realIP != "" {
		req.Header.Add("X-Real-IP", realIP)
	}

	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Printf("status: %d %s body: %s\n", resp.StatusCode, resp.Status, string(respBody))

	return resp.StatusCode, string(respBody)
}

func httpRequest(ts *httptest.Server, method, path string, body []byte) (int, string) {
	req, _ := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func httpRequestGzip(ts *httptest.Server, method, path string, body []byte) (int, string) {
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
	var c int64
	if !counterType(c) {
		t.Errorf("Error: Variable is counter type.")
	}

	var g float64
	if !gaugeType(g) {
		t.Errorf("Error: Variable is gauge type.")
	}
}

func TestContentEncodingContains(t *testing.T) {
	contentEncodingValues := []string{"deflate", "gzip", "bzip"}

	if !contentEncodingContains(contentEncodingValues, "deflate") {
		t.Errorf("Error: content contain deflate")
	}

	if !contentEncodingContains(contentEncodingValues, "gzip") {
		t.Errorf("Error: content contain gzip")
	}

	if contentEncodingContains(contentEncodingValues, "fake") {
		t.Errorf("Error: content not contain fake")
	}
}

func NewTestServer(ctx context.Context, cfg *Config) *HTTPServer {
	s := NewServer(ctx, cfg)

	s.chiRouter.Use(middleware.RealIP)
	s.chiRouter.Use(s.decryptRSABodyHandler)
	s.chiRouter.Use(unzipBodyHandler)
	s.chiRouter.Get("/", s.listHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Route("/", func(r chi.Router) {
		r.Use(s.trustedSubnetHandler)
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/value/", s.GetValueHandler)
		r.Post("/update/", s.UpdateHandler)
		r.Post("/updates/", s.UpdatesHandler)
	})

	/*
		s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))
		s.chiRouter.Use(unzipBodyHandler)
		s.chiRouter.Get("/", s.listHandler)
		s.chiRouter.Post("/", s.defaultHandler)
		s.chiRouter.Route("/", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))
			r.Post("/value/", s.GetValueHandler)
			r.Post("/update/", s.UpdateHandler)
			r.Post("/updates/", s.UpdatesHandler)
		})
	*/

	return s
}

// test server without hash
func TestHTTPServer1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	s := NewTestServer(ctx, cfg)

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()

	tests := []struct {
		name               string
		action             string
		requestBody        []byte
		responseStatusCode int
		responseBody       string
	}{
		{
			name:               "update gauge #1",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update gauge #2",
			action:             "update",
			requestBody:        []byte(`{"id":"BuckHashSys","type":"gauge","value":3877}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update counter #3",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","delta":1}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update non gauge value #4",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","delta":1}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update non counter value #5",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","value":2}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update unsupport type #6",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"integer","value":2}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update empty body #7",
			action:             "update",
			requestBody:        []byte(`{}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update empty body #8",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount"}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "get value gauge #1",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"Alloc","type":"gauge","value":1907608}`,
		},
		{
			name:               "get value counter #2",
			action:             "value",
			requestBody:        []byte(`{"id":"PollCount","type":"counter"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"PollCount","type":"counter","delta":2}`,
		},
		{
			name:               "get unknown value #3",
			action:             "value",
			requestBody:        []byte(`{"id":"New","type":"counter"}`),
			responseStatusCode: http.StatusNotFound,
			responseBody:       "",
		},
		{
			name:               "get unknown value #4",
			action:             "value",
			requestBody:        []byte(`{}`),
			responseStatusCode: http.StatusNotFound,
			responseBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "update" {
				statusCode, body := httpRequest(ts, "POST", "/update/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}

				statusCode, body = httpRequestGzip(ts, "POST", "/update/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			if tt.action == "value" {
				statusCode, body := httpRequest(ts, "POST", "/value/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}

				statusCode, body = httpRequestGzip(ts, "POST", "/value/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}
		})
	}
}

// run test with postgres db
func TestHTTPServerWithDB(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	cfg.DatabaseDSN = "postgresql://postgres:postgres@postgres:5432/praktikum"
	s := NewTestServer(ctx, cfg)
	s.Storage.(*postgres.Postgres).InitDB()
	s.Storage.(*postgres.Postgres).Clear()

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()
	defer s.Storage.(*postgres.Postgres).Close()

	tests := []struct {
		name               string
		action             string
		requestBody        []byte
		responseStatusCode int
		responseBody       string
	}{
		{
			name:               "update gauge #1",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update gauge #2",
			action:             "update",
			requestBody:        []byte(`{"id":"BuckHashSys","type":"gauge","value":3877}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update counter #3",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","delta":1}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update non gauge value #4",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","delta":1}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update non counter value #5",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","value":2}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update unsupport type #6",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"integer","value":2}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update empty body #7",
			action:             "update",
			requestBody:        []byte(`{}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update empty body #8",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount"}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "get value gauge #1",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"Alloc","type":"gauge","value":1907608}`,
		},
		{
			name:               "get value counter #2",
			action:             "value",
			requestBody:        []byte(`{"id":"PollCount","type":"counter"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"PollCount","type":"counter","delta":1}`,
		},
		{
			name:               "get unknown value #3",
			action:             "value",
			requestBody:        []byte(`{"id":"New","type":"counter"}`),
			responseStatusCode: http.StatusNotFound,
			responseBody:       "",
		},
		{
			name:               "get unknown value #4",
			action:             "value",
			requestBody:        []byte(`{}`),
			responseStatusCode: http.StatusNotFound,
			responseBody:       "",
		},
		{
			name:   "updates multi data #1",
			action: "updates",
			requestBody: []byte(`[{"id":"PollCount1","type":"counter","delta":1},
                                        {"id":"PollCount2","type":"counter","delta":1},
										{"id":"Alloc1","type":"gauge","value":1907608},
                                        {"id":"Alloc2","type":"gauge","value":1907777}]`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "get value after multi data #2",
			action:             "value",
			requestBody:        []byte(`{"id":"PollCount1","type":"counter"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"PollCount1","type":"counter","delta":1}`,
		},
		{
			name:               "get value after multi data #3",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc2","type":"gauge"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"Alloc2","type":"gauge","value":1907777}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set single data
			if tt.action == "update" {
				statusCode, body := httpRequest(ts, "POST", "/update/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			// set multi data
			if tt.action == "updates" {
				statusCode, body := httpRequest(ts, "POST", "/updates/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			if tt.action == "value" {
				statusCode, body := httpRequest(ts, "POST", "/value/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}
		})
	}
}

// test server with hash
func TestHTTPServerHash(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	cfg.KeySign = "secret"

	s := NewTestServer(ctx, cfg)

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()

	tests := []struct {
		name               string
		action             string
		requestBody        []byte
		responseStatusCode int
		responseBody       string
	}{
		{
			name:               "update gauge #1",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":226640,"hash":"3544777d62d524efaacb5eae93073cb716251bff20490e6e5c266376dc002f3e"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "update counter #2",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","delta":1,"hash":"ce97c6062da4477a5fad4cfdd24f0f24e474d309b1f054928dd138683d1cab12"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "get value counter #3",
			action:             "value",
			requestBody:        []byte(`{"id":"PollCount","type":"counter"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"PollCount","type":"counter","delta":1,"hash":"ce97c6062da4477a5fad4cfdd24f0f24e474d309b1f054928dd138683d1cab12"}`,
		},
		{
			name:               "get value gauge #4",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"Alloc","type":"gauge","value":226640,"hash":"3544777d62d524efaacb5eae93073cb716251bff20490e6e5c266376dc002f3e"}`,
		},
		{
			name:               "update incorrect hash #5",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":226640,"hash":"3544777d62d524efaacb5eae93073cb716251bff20490e6e5c266376dc000000"}`),
			responseStatusCode: http.StatusBadRequest,
			responseBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "update" {
				statusCode, body := httpRequest(ts, "POST", "/update/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			if tt.action == "value" {
				statusCode, body := httpRequest(ts, "POST", "/value/", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}
		})
	}
}

// test server with X-Real-IP
func TestHTTPServerRealIP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	cfg.TrustedSubnet = "192.168.0.0/24"

	s := NewTestServer(ctx, cfg)

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()

	tests := []struct {
		name               string
		action             string
		requestBody        []byte
		realIP             string
		responseStatusCode int
		responseBody       string
	}{
		{
			name:               "update gauge from empty x-real-ip #1",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			realIP:             "",
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update gauge from not trusted ip #2",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			realIP:             "192.168.1.1",
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "update gauge from trusted ip #3",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			realIP:             "192.168.0.1",
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:               "get value gauge from empty x-real-ip #1",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			realIP:             "",
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "get value gauge from not trusted ip #2",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			realIP:             "192.168.1.1",
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:               "get value gauge from trusted ip #2",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			realIP:             "192.168.0.1",
			responseStatusCode: http.StatusOK,
			responseBody:       `{"id":"Alloc","type":"gauge","value":1907608}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "update" {
				statusCode, body := httpRequestRealIP(ts, "POST", "/update/", tt.requestBody, tt.realIP)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			if tt.action == "value" {
				statusCode, body := httpRequestRealIP(ts, "POST", "/value/", tt.requestBody, tt.realIP)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}
		})
	}
}
