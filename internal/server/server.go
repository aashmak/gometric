package server

import (
	"context"
	"gometric/internal/storage"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func NewServer(ctx context.Context, cfg *Config) *HTTPServer {
	httpserver := HTTPServer{
		Storage:   storage.New(),
		chiRouter: chi.NewRouter(),
	}

	httpserver.Storage.SetStoreFile(cfg.StoreFile)

	if cfg.StoreInterval > 0 {
		httpserver.StoreHandler(ctx, cfg.StoreInterval)
	} else {
		httpserver.Storage.SetSyncMode(true)
	}

	httpserver.Storage.Open()

	return &httpserver
}

func (s HTTPServer) Restore() error {
	data, err := s.Storage.LoadDump()
	if err != nil {
		return err
	}

	for k, v := range data {
		switch k {
		case "PollCount":
			if err := s.Storage.Set(k, counter(v.(float64))); err != nil {
				return err
			}
		default:
			if err := s.Storage.Set(k, gauge(v.(float64))); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s HTTPServer) StoreHandler(ctx context.Context, storeInterval int) {
	go func(ctx context.Context, storeInterval int) {
		var interval = time.Duration(storeInterval) * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := s.Storage.SaveDump(); err != nil {
					log.Print(err)
				}

				<-time.After(interval)
			}
		}

	}(ctx, storeInterval)
}

func (s *HTTPServer) ListenAndServe(addr string) {

	// middleware gzip response
	s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))

	// middleware unzip request
	s.chiRouter.Use(unzipBodyHandler)
	s.chiRouter.Get("/", s.listHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Post("/value/", s.GetValueHandler)
	s.chiRouter.Post("/update/", s.UpdateHandler)

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
		log.Fatalf("Server shutdown failed:%+v", err)
	}

	if err := s.Storage.Close(); err != nil {
		log.Fatalf("Server storage close is failed:%+v", err)
	}
}
