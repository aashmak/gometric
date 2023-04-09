package agent

import (
	"encoding/json"
	"testing"
)

func TestCollectorRegisterMetric(t *testing.T) {
	m := &MemStats{
		Alloc:     1.0000,
		PollCount: 1,
	}

	collector := Collector{}

	// register new gauge metric
	err := collector.RegisterMetric("Alloc", &(m.Alloc))
	if err != nil || ((float64)(*collector.Metrics[0].Value) != 1.0000 && (string)(collector.Metrics[0].MType) != "gauge") {
		t.Errorf("error")
	}

	// register exist metric
	(*m).Alloc = 2.0000
	err = collector.RegisterMetric("Alloc", &(m.Alloc))
	if err == nil || (float64)(*collector.Metrics[0].Value) != 2.0000 {
		t.Errorf("error")
	}

	// register new counter metric
	err = collector.RegisterMetric("PollCount", &(m.PollCount))
	if err != nil || ((int64)(*collector.Metrics[1].Delta) != 1 && (string)(collector.Metrics[1].MType) != "counter") {
		t.Errorf("error")
	}

	// test to Marshal
	jsonTestStr := "[{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":2},{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":1}]"

	ret, err := json.Marshal(collector.Metrics)
	if err != nil || string(ret) != jsonTestStr {
		t.Errorf("error")
	}
}

func TestReadMemStats(t *testing.T) {
	var m MemStats
	m.ReadMemStats()

	if m.Alloc == 0 {
		t.Errorf("Alloc expected to be float")
	}

	if m.PollCount != 1 {
		t.Errorf("PollCount expected to be 1")
	}
}
