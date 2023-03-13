package metric

import "testing"

func TestCollectorRegisterMetric(t *testing.T) {
	m := &MemStats{
		Alloc:     1.0000,
		PollCount: 1,
	}

	collector := Collector{}

	// register new gauge metric
	err := collector.RegisterMetric("Alloc", &(m.Alloc))
	if err != nil || *collector.Metrics["Alloc"].(*gauge) != 1.0000 {
		t.Errorf("error")
	}

	// register exist metric
	(*m).Alloc = 2.0000
	err = collector.RegisterMetric("Alloc", &(m.Alloc))
	if err == nil || *collector.Metrics["Alloc"].(*gauge) != 2.0000 {
		t.Errorf("error")
	}

	// register new counter metric
	err = collector.RegisterMetric("PollCount", &(m.PollCount))
	if err != nil || *collector.Metrics["PollCount"].(*counter) != 1 {
		t.Errorf("error")
	}

	// register new int metric
	var new_metric int
	new_metric = 1
	err = collector.RegisterMetric("New", &new_metric)
	if err == nil {
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
