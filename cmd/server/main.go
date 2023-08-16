package main

import (
	"context"
	"errors"
	"fmt"
	"gometric/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env"
	"github.com/jessevdk/go-flags"
)

func main() {
	cfg := server.DefaultConfig()

	parser := flags.NewParser(cfg, flags.HelpFlag)
	if _, err := parser.Parse(); err != nil {
		var e *flags.Error

		if errors.As(err, &e) {
			if e.Type == flags.ErrHelp {
				fmt.Printf("%s", e.Message)
				os.Exit(0)
			}
		}
		log.Fatalf("error parse arguments:%+v\n", err)
	}

	//The values of the config is overridden
	//from the environment variables
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Error parse environment:%+v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
