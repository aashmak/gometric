package metric

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Collector struct {
	Endpoint          string
	ReportIntervalSec int
	Metrics           map[string]interface{}
}

func (c *Collector) RegisterMetric(name string, value interface{}) error {
	if c.Metrics == nil {
		c.Metrics = make(map[string]interface{})
	}

	if _, ok := c.Metrics[name]; ok {
		return errors.New(fmt.Sprintf("Metric %s already exists", name))
	}

	switch value.(type) {
	case *gauge:
		c.Metrics[name] = value.(*gauge)
	case *counter:
		c.Metrics[name] = value.(*counter)
	default:
		return errors.New(fmt.Sprintf("Unknown metric type"))
	}

	return nil
}

func (c *Collector) SendMetric() {
	var interval = time.Duration(c.ReportIntervalSec) * time.Second
	client := &http.Client{}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		defer cancel()

		select {
		case <-ctx.Done():
			continue

		default:
			fmt.Println("work...")
			go func() {
				for key, value := range c.Metrics {
					switch c.Metrics[key].(type) {
					case *counter:
						url := fmt.Sprintf("%s/counter/%s/%d", c.Endpoint, key, *value.(*counter))
						fmt.Println(url)
						MakeRequest(ctx, client, url)
					case *gauge:
						url := fmt.Sprintf("%s/gauge/%s/%.4f", c.Endpoint, key, *value.(*gauge))
						time.Sleep(2 * time.Second)
						fmt.Println(url)
						MakeRequest(ctx, client, url)
					}
				}
			}()
		}

		<-time.After(interval)
		cancel()
	}
}

func MakeRequest(ctx context.Context, client *http.Client, url string) error {
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		return errors.New(fmt.Sprintf("Http request error: %s", err.Error()))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil
	}

	return nil
}
