// Пакет server предназначен для организации http сервера с REST API для сбора метрик.
package server

import (
	"context"
	"crypto/rsa"
	"net/http"
	"time"

	"gometric/internal/crypto"
	"gometric/internal/logger"
	"gometric/internal/memstorage"
	"gometric/internal/postgres"
	"gometric/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HTTPServer описывает структуру сервера.
type HTTPServer struct {
	Server        *http.Server
	chiRouter     chi.Router
	Storage       storage.Storage
	KeySign       string
	RSAPrivateKey *rsa.PrivateKey
}

// NewServer создает новый http сервер.
func NewServer(ctx context.Context, cfg *Config) *HTTPServer {
	httpserver := HTTPServer{
		chiRouter: chi.NewRouter(),
		KeySign:   cfg.KeySign,
	}

	if cfg.RSAPrivateKey != "" {
		var err error
		httpserver.RSAPrivateKey, err = crypto.NewPrivateKey(cfg.RSAPrivateKey)
		if err != nil {
			logger.Fatal("new private key failed", err)
		}
		logger.Debug("new private key initialized")
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

// Restore загрузка данных из файла в in-memory БД.
// Используется только для бэкенда MemStorageDB.
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

// StoreHandler сохранение данных из in-memory БД в файл.
// Используется только для бэкенда MemStorageDB.
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

// ListenAndServe старт сервера
func (s *HTTPServer) ListenAndServe(addr string) {

	// middleware gzip response
	// s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))

	// middleware decrypt body
	s.chiRouter.Use(s.decryptRSABodyHandler)
	// middleware unzip body
	s.chiRouter.Use(unzipBodyHandler)

	s.chiRouter.Get("/", s.listHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/value/", s.GetValueHandler)
		r.Post("/update/", s.UpdateHandler)
		r.Post("/updates/", s.UpdatesHandler)
	})
	s.chiRouter.Get("/ping", s.pingHandler)
	// s.chiRouter.Post("/value/", s.GetValueHandler)
	// s.chiRouter.Post("/update/", s.UpdateHandler)
	// s.chiRouter.Post("/updates/", s.UpdatesHandler)
	s.chiRouter.Mount("/debug", middleware.Profiler())

	s.Server = &http.Server{
		Addr:    addr,
		Handler: s.chiRouter,
	}

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("", err)
	}
}

// Shutdown завершение работы сервера
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
