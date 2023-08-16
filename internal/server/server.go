package server

import (
	"context"
	"gometric/internal/memstorage"
	"gometric/internal/postgres"
	"gometric/internal/storage"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type gauge float64
type counter int64

type HTTPServer struct {
	Server    *http.Server
	chiRouter chi.Router
	Storage   storage.Storage
	KeySign   []byte
}

func NewServer(ctx context.Context, cfg *Config) *HTTPServer {
	httpserver := HTTPServer{
		chiRouter: chi.NewRouter(),
		KeySign:   []byte(cfg.KeySign),
	}

	if cfg.DatabaseDSN != "" {
		httpserver.Storage = storage.NewPostgresDB(cfg.DatabaseDSN)
		httpserver.Storage.Open()
		err := httpserver.Storage.(*postgres.Postgres).InitDB()
		if err != nil {
			log.Fatalf("Init db error: %s", err.Error())
		}
	} else {
		syncMode := false

		if cfg.StoreInterval == 0 {
			syncMode = true
		}

		httpserver.Storage = storage.NewMemStorage(cfg.StoreFile, syncMode)
		httpserver.Storage.Open()
		httpserver.StoreHandler(ctx, cfg.StoreInterval)

		if cfg.Restore {
			if err := httpserver.Restore(); err != nil {
				log.Print(err)
			}
		}
	}

	return &httpserver
}

// restore MemStorageDB from file
func (s HTTPServer) Restore() error {
	data, err := s.Storage.(*memstorage.MemStorage).LoadDump()
	if err != nil {
		return err
	}

	for k, v := range data {
		switch k {
		case "PollCount":
			if err := s.Storage.Set(k, int64(v.(float64))); err != nil {
				return err
			}
		default:
			if err := s.Storage.Set(k, v.(float64)); err != nil {
				return err
			}
		}
	}

	return nil
}

// save MemStorageDB to file
func (s HTTPServer) StoreHandler(ctx context.Context, storeInterval int) {
	if storeInterval <= 0 {
		return
	}

	go func(ctx context.Context, storeInterval int) {
		var interval = time.Duration(storeInterval) * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := s.Storage.(*memstorage.MemStorage).SaveDump(); err != nil {
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
	s.chiRouter.Get("/ping", s.pingHandler)
	s.chiRouter.Post("/value/", s.GetValueHandler)
	s.chiRouter.Post("/update/", s.UpdateHandler)
	s.chiRouter.Post("/updates/", s.UpdatesHandler)

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
