package main

import (
	"context"
	"gometric/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v7"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := server.DefaultConfig()
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Error parse environment:%+v", err)
	}

	serv := server.NewServer(ctx, cfg)

	if cfg.Restore {
		if err := serv.Restore(); err != nil {
			log.Print(err)
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	go serv.ListenAndServe(cfg.ListenAddr)
	log.Print("Server started")

	<-sigint

	serv.Shutdown()
	log.Print("Server stopped")
}
