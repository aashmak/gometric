package metric

import (
	"context"
	"fmt"
	"log"
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
		return fmt.Errorf("metric %s already exists", name)
	}

	switch v := value.(type) {
	case *gauge:
		c.Metrics[name] = value.(*gauge)
	case *counter:
		c.Metrics[name] = value.(*counter)
	default:
		return fmt.Errorf("unknown metric type %v", v)
	}

	return nil
}

func (c *Collector) SendMetric(ctx context.Context) {

	var interval = time.Duration(c.ReportIntervalSec) * time.Second
	client := &http.Client{}

	for {
		ctxSendMetric, cancel := context.WithTimeout(ctx, interval)
		defer cancel()

		select {
		case <-ctxSendMetric.Done():
			continue
		case <-ctx.Done():
			log.Print("SendMetric stopped")
			return

		default:
			go func() {
				var url string

				for key, value := range c.Metrics {
					switch c.Metrics[key].(type) {
					case *counter:
						url = fmt.Sprintf("%s/counter/%s/%d", c.Endpoint, key, *value.(*counter))
					case *gauge:
						url = fmt.Sprintf("%s/gauge/%s/%.4f", c.Endpoint, key, *value.(*gauge))
					default:
						continue
					}

					err := MakeRequest(ctxSendMetric, client, url)
					if err != nil {
						log.Printf("Http request error: %s", err.Error())
					}
				}
			}()
		}

		<-time.After(interval)
	}
}

func MakeRequest(ctx context.Context, client *http.Client, url string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("new request error: %s", err.Error())
	}

	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("http request error: %s", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("the request was not executed successfully")
	}

	return nil
}
