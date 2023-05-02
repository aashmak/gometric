package agent

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"gometric/internal/logger"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type gauge float64
type counter int64

type MemStats struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge
	PollCount     counter
	RandomValue   gauge
}

func (m *MemStats) ReadMemStats() {
	var mstat runtime.MemStats
	runtime.ReadMemStats(&mstat)

	m.Alloc = gauge(mstat.Alloc)
	m.BuckHashSys = gauge(mstat.BuckHashSys)
	m.Frees = gauge(mstat.Frees)
	m.GCCPUFraction = gauge(mstat.GCCPUFraction)
	m.GCSys = gauge(mstat.GCSys)
	m.HeapAlloc = gauge(mstat.HeapAlloc)
	m.HeapIdle = gauge(mstat.HeapIdle)
	m.HeapInuse = gauge(mstat.HeapInuse)
	m.HeapObjects = gauge(mstat.HeapObjects)
	m.HeapReleased = gauge(mstat.HeapReleased)
	m.HeapSys = gauge(mstat.HeapSys)
	m.LastGC = gauge(mstat.LastGC)
	m.Lookups = gauge(mstat.Lookups)
	m.MCacheInuse = gauge(mstat.MCacheInuse)
	m.MCacheSys = gauge(mstat.MCacheSys)
	m.MSpanInuse = gauge(mstat.MSpanInuse)
	m.MSpanSys = gauge(mstat.MSpanSys)
	m.Mallocs = gauge(mstat.Mallocs)
	m.NextGC = gauge(mstat.NextGC)
	m.NumForcedGC = gauge(mstat.NumForcedGC)
	m.NumGC = gauge(mstat.NumGC)
	m.OtherSys = gauge(mstat.OtherSys)
	m.PauseTotalNs = gauge(mstat.PauseTotalNs)
	m.StackInuse = gauge(mstat.StackInuse)
	m.StackSys = gauge(mstat.StackSys)
	m.Sys = gauge(mstat.Sys)
	m.TotalAlloc = gauge(mstat.TotalAlloc)
	m.PollCount++
	m.RandomValue = gauge(rand.Float64())
}

type VirtualMemoryStat struct {
	Total gauge
	Free  gauge
}

func (v *VirtualMemoryStat) VirtualMemory() error {
	vmemstat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	v.Total = gauge(vmemstat.Total)
	v.Free = gauge(vmemstat.Free)

	return nil
}

type CPUStat struct {
	Counts  int
	Percent []gauge
}

func (c *CPUStat) CPU(ctx context.Context) error {
	var percents []float64
	var err error

	percents, err = cpu.PercentWithContext(ctx, 10, true)
	if err != nil {
		return err
	}

	percentsNew := make([]gauge, len(percents))
	for percent := range percents {
		percentsNew = append(percentsNew, gauge(percent))
	}

	c.Percent = percentsNew

	return nil
}

func MemStatCollector(ctx context.Context, pollInterval int) *MemStats {
	var m MemStats

	go func(ctx context.Context, m *MemStats, pollInterval int) {
		var interval = time.Duration(pollInterval) * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				m.ReadMemStats()
				<-time.After(interval)
			}
		}
	}(ctx, &m, pollInterval)

	return &m
}

func VirtualMemoryCollector(ctx context.Context, pollInterval int) *VirtualMemoryStat {
	var v VirtualMemoryStat
	var err error

	go func(ctx context.Context, v *VirtualMemoryStat, pollInterval int) {
		var interval = time.Duration(pollInterval) * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err = v.VirtualMemory()
				if err != nil {
					logger.Error("", err)
				}

				<-time.After(interval)
			}
		}
	}(ctx, &v, pollInterval)

	return &v
}

func CPUCollector(ctx context.Context, pollInterval int) *CPUStat {
	var c CPUStat
	var err error

	c.Counts, err = cpu.CountsWithContext(ctx, true)
	if err != nil {
		return nil
	}

	c.Percent = make([]gauge, c.Counts)

	go func(ctx context.Context, c *CPUStat, pollInterval int) {
		var interval = time.Duration(pollInterval) * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err = c.CPU(ctx)
				if err != nil {
					logger.Error("", err)
				}

				<-time.After(interval)
			}
		}
	}(ctx, &c, pollInterval)

	return &c
}
