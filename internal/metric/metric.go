package metric

import (
	"math/rand"
	"runtime"
	"time"
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

func RunCollector(pollInterval int) *MemStats {
	var m *MemStats
	var c chan *MemStats = make(chan *MemStats)

	go RunCollectorHandler(c, pollInterval)
	m = <-c

	return m
}

func RunCollectorHandler(c chan *MemStats, pollInterval int) {
	var interval = time.Duration(pollInterval) * time.Second
	var m MemStats

	c <- &m
	close(c)

	for {
		m.ReadMemStats()
		<-time.After(interval)
	}
}
