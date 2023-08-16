package main

import (
	"gometric/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	ListenAddr string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error parse environment:%+v", err)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	serv := server.NewServer()
	go serv.ListenAndServe(cfg.ListenAddr)
	log.Print("Server started")

	<-sigint

	serv.Shutdown()
	log.Print("Server stopped")
}
