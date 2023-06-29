package server

import (
	"context"
	"gometric/internal/logger"
	"gometric/internal/memstorage"
	"gometric/internal/postgres"
	"gometric/internal/storage"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
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
		var db *pgxpool.Pool

		poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("unable to parse database dsn", err)
		}

		db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			logger.Fatal("unable to create connection pool", err)
		}

		httpserver.Storage = storage.NewPostgresDB(db)
		err = httpserver.Storage.(*postgres.Postgres).InitDB()
		if err != nil {
			logger.Fatal("", err)
		}
	} else {
		syncMode := false

		if cfg.StoreInterval == 0 {
			syncMode = true
		}

		var err error
		httpserver.Storage, err = storage.NewMemStorage(cfg.StoreFile, syncMode)
		if err != nil {
			logger.Fatal("new MemStorage", err)
		}

		httpserver.StoreHandler(ctx, cfg.StoreInterval)

		if cfg.Restore {
			if err := httpserver.Restore(); err != nil {
				logger.Error("restore MemStorage", err)
			}
		}
	}

	return &httpserver
}

// restore MemStorageDB from file
func (s HTTPServer) Restore() error {
	data, err := s.Storage.(*memstorage.MemStorage).LoadDump()
	if err != nil {
		logger.Error("restore from json db", err)
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
					logger.Error("save dump to json db", err)
				}

				<-time.After(interval)
			}
		}

	}(ctx, storeInterval)
}

func (s *HTTPServer) ListenAndServe(addr string) {

	// middleware gzip response
	//s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))

	// middleware unzip request
	s.chiRouter.Use(unzipBodyHandler)

	s.chiRouter.Get("/", s.listHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Get("/ping", s.pingHandler)
	s.chiRouter.Post("/value/", s.GetValueHandler)
	s.chiRouter.Post("/update/", s.UpdateHandler)
	//s.chiRouter.Post("/updates/", s.UpdatesHandler)
	s.chiRouter.Mount("/debug", middleware.Profiler())

	s.Server = &http.Server{
		Addr:    addr,
		Handler: s.chiRouter,
	}

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("", err)
	}
}

func (s *HTTPServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed", err)
	}

	if err := s.Storage.Close(); err != nil {
		logger.Fatal("Server storage close is failed", err)
	}
}
