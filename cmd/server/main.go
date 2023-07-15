package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gometric/internal/logger"
	"gometric/internal/server"

	"github.com/caarlos0/env"
	"github.com/jessevdk/go-flags"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	cfg := server.DefaultConfig()

	parser := flags.NewParser(cfg, flags.HelpFlag)
	if _, err := parser.Parse(); err != nil {
		var e *flags.Error

		if errors.As(err, &e) {
			if e.Type == flags.ErrHelp {
				log.Printf("%s", e.Message)
				exit(0)
			}
		}
		log.Fatalf("error parse arguments:%+v\n", err)
	}

	// The values of the config is overridden
	// from the environment variables
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Error parse environment:%+v\n", err)
	}

	// Print version
	if cfg.Version {
		printVersion()
		exit(0)
	}

	logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	defer logger.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	serv := server.NewServer(ctx, cfg)
	go serv.ListenAndServe(cfg.ListenAddr)

	printVersion()
	logger.Info("Server started")

	<-sigint

	serv.Shutdown()
	logger.Info("Server stopped")
}

func exit(code int) {
	os.Exit(code)
}

func printVersion() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}
