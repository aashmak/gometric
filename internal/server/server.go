package server

import (
	"context"
	"encoding/json"
	"fmt"
	"gometric/internal/storage"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type gauge float64
type counter int64

type HTTPServer struct {
	Server    *http.Server
	chiRouter chi.Router
	Storage   storage.Storage
}

func NewServer() *HTTPServer {
	return &HTTPServer{
		Storage:   storage.New(),
		chiRouter: chi.NewRouter(),
	}
}

func (s HTTPServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
}

func (s HTTPServer) listHandler(w http.ResponseWriter, r *http.Request) {
	var varList string

	for _, metricName := range s.Storage.List() {
		v, err := s.Storage.Get(metricName)
		if err == nil {
			if gaugeType(v) {
				varList += fmt.Sprintf("%s (type: gauge): %f<br>\n", metricName, v.(gauge))
			} else if counterType(v) {
				varList += fmt.Sprintf("%s (type: counter): %d<br>\n", metricName, v.(counter))
			}
		}
	}

	fmt.Fprintf(w, "<html>\n<title>Metric Dump</title>\n"+
		"<body>\n<h2>Metric Dump</h2>\n"+
		"%s\n"+
		"</body>\n</html>", varList)
}

func (s HTTPServer) GetValueHandler(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") == "application/json" {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("server: could not read request body: %s\n", err)
		}

		var metric Metrics
		if err := json.Unmarshal(reqBody, &metric); err != nil {
			log.Printf("Error: %s", err.Error())
			return
		}

		v, err := s.Storage.Get(metric.ID)
		if err == nil {
			switch metric.MType {
			case "gauge":
				if gaugeType(v) {
					v1 := v.(gauge)
					metric.Value = (*float64)(&v1)

					ret, err := json.Marshal(v)
					if err != nil {
						log.Printf("Error: %s", err.Error())
					}
					w.Write(ret)
					return
				}
			case "counter":
				if counterType(v) {
					v1 := v.(counter)
					metric.Delta = (*int64)(&v1)

					ret, err := json.Marshal(v)
					if err != nil {
						log.Printf("Error: %s", err.Error())
					}
					w.Write(ret)
					return
				}
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s HTTPServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
	}

	var metric Metrics
	if err := json.Unmarshal(reqBody, &metric); err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	switch metric.MType {
	case "gauge":
		if metric.ID != "" && metric.Value != nil {
			err := s.Storage.Set(metric.ID, gauge(*metric.Value))
			if err == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	case "counter":
		// get previous counter value
		prevCounter, err := s.Storage.Get(metric.ID)
		if err != nil {
			prevCounter = counter(0)
		}

		if metric.ID != "" && metric.Delta != nil {
			err = s.Storage.Set(metric.ID, counter(*metric.Delta)+prevCounter.(counter))
			if err == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	w.WriteHeader(http.StatusForbidden)
}

func (s *HTTPServer) ListenAndServe(addr string) {

	s.chiRouter.Route("/", func(router chi.Router) {
		s.chiRouter.Get("/", s.listHandler)
		s.chiRouter.Post("/", s.defaultHandler)
		s.chiRouter.Post("/value/", s.GetValueHandler)
		s.chiRouter.Post("/update/", s.UpdateHandler)
	})

	s.Server = &http.Server{
		Addr:    addr,
		Handler: s.chiRouter,
	}

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func (s *HTTPServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}
