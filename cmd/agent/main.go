package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gometric/internal/agent"
	"gometric/internal/logger"

	"github.com/caarlos0/env/v7"
	"github.com/jessevdk/go-flags"
)

type Config struct {
	EndpointAddr   string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set remote metric collector"`
	ReportInterval int    `long:"report_interval" short:"r" env:"REPORT_INTERVAL" default:"10" description:"set report interval"`
	PollInterval   int    `long:"poll_interval" short:"p" env:"POLL_INTERVAL" default:"2" description:"set poll interval"`
	KeySign        string `long:"key" short:"k" env:"KEY" description:"set key for signing"`
	LogLevel       string `long:"log_level" env:"LOG_LEVEL" default:"info" description:"set log level"`
	LogFile        string `long:"log_file" env:"LOG_FILE" default:"" description:"set log file"`
	RateLimit      int    `long:"rate_limit" short:"l" env:"RATE_LIMIT" default:"1" description:"set rate limit"`
	Version        bool   `long:"version" short:"v" description:"print current version"`
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	var cfg Config

	parser := flags.NewParser(&cfg, flags.HelpFlag)
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
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("error parse env variables:%+v\n", err)
	}

	// Print version
	if cfg.Version {
		printVersion()
		exit(0)
	}

	// Init logger
	logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	defer logger.Close()

	printVersion()
	logger.Info("Agent started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run metric collector with pollInterval 2 sec
	m := agent.MemStatCollector(ctx, cfg.PollInterval)
	v := agent.VirtualMemoryCollector(ctx, cfg.PollInterval)
	c := agent.CPUCollector(ctx, cfg.PollInterval)

	collector := agent.Collector{
		Endpoint:          "http://" + cfg.EndpointAddr + "/update/",
		ReportIntervalSec: cfg.ReportInterval,
		KeySign:           cfg.KeySign,
		RateLimit:         cfg.RateLimit,
	}

	// runtime metrics
	collector.RegisterMetric("Alloc", &(m.Alloc))
	collector.RegisterMetric("BuckHashSys", &(m.BuckHashSys))
	collector.RegisterMetric("Frees", &(m.Frees))
	collector.RegisterMetric("GCCPUFraction", &(m.GCCPUFraction))
	collector.RegisterMetric("GCSys", &(m.GCSys))
	collector.RegisterMetric("HeapAlloc", &(m.HeapAlloc))
	collector.RegisterMetric("HeapIdle", &(m.HeapIdle))
	collector.RegisterMetric("HeapInuse", &(m.HeapInuse))
	collector.RegisterMetric("HeapObjects", &(m.HeapObjects))
	collector.RegisterMetric("HeapReleased", &(m.HeapReleased))
	collector.RegisterMetric("HeapSys", &(m.HeapSys))
	collector.RegisterMetric("LastGC", &(m.LastGC))
	collector.RegisterMetric("Lookups", &(m.Lookups))
	collector.RegisterMetric("MCacheInuse", &(m.MCacheInuse))
	collector.RegisterMetric("MCacheSys", &(m.MCacheSys))
	collector.RegisterMetric("MSpanInuse", &(m.MSpanInuse))
	collector.RegisterMetric("MSpanSys", &(m.MSpanSys))
	collector.RegisterMetric("Mallocs", &(m.Mallocs))
	collector.RegisterMetric("NextGC", &(m.NextGC))
	collector.RegisterMetric("NumForcedGC", &(m.NumForcedGC))
	collector.RegisterMetric("NumGC", &(m.NumGC))
	collector.RegisterMetric("OtherSys", &(m.OtherSys))
	collector.RegisterMetric("PauseTotalNs", &(m.PauseTotalNs))
	collector.RegisterMetric("StackInuse", &(m.StackInuse))
	collector.RegisterMetric("StackSys", &(m.StackSys))
	collector.RegisterMetric("Sys", &(m.Sys))
	collector.RegisterMetric("TotalAlloc", &(m.TotalAlloc))
	collector.RegisterMetric("PollCount", &(m.PollCount))
	collector.RegisterMetric("RandomValue", &(m.RandomValue))

	// virtual memory metrics
	collector.RegisterMetric("TotalMemory", &(v.Total))
	collector.RegisterMetric("FreeMemory", &(v.Free))

	// cpu utilization metrics
	for i := 0; i < c.Counts; i++ {
		collector.RegisterMetric(fmt.Sprintf("CPUutilization%d", i+1), &(c.Percent[i]))
	}

	// send metrics
	go collector.SendMetric(ctx)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	<-sigint

	logger.Info("Agent stopped")
}

func exit(code int) {
	os.Exit(code)
}

func printVersion() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}
