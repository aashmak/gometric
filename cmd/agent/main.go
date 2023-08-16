package main

import (
	"context"
	"gometric/internal/metric"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	EndpointAddr   string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error parse environment:%+v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Run metric collector with pollInterval 2 sec
	m := metric.RunCollector(ctx, cfg.PollInterval)

	collector := metric.Collector{
		Endpoint:          "http://" + cfg.EndpointAddr + "/update/",
		ReportIntervalSec: cfg.ReportInterval,
	}

	//prepare metrics for collector
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

	//send metrics
	go collector.SendMetric(ctx)
	log.Print("Agent started")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	<-sigint

	log.Print("Agent stopped")
}
