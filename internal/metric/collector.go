package metric

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Collector struct {
	Endpoint          string
	ReportIntervalSec int
	Metrics           []Metrics
}

func (c *Collector) RegisterMetric(name string, value interface{}) error {
	if c.Metrics == nil {
		c.Metrics = make([]Metrics, 0)
	}

	for _, v := range c.Metrics {
		if v.ID == name {
			return fmt.Errorf("metric %s already exists", name)
		}
	}

	tmp := Metrics{
		ID: name,
	}

	switch v := value.(type) {
	case *gauge:
		tmp.MType = "gauge"
		tmp.Value = (*float64)(value.(*gauge))
	case *counter:
		tmp.MType = "counter"
		tmp.Delta = (*int64)(value.(*counter))
	default:
		return fmt.Errorf("unknown metric type %v", v)
	}

	c.Metrics = append(c.Metrics, tmp)

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
				for _, v := range c.Metrics {

					ret, err := json.Marshal(v)
					if err != nil {
						log.Printf("Error: %s", err.Error())
					}

					err = MakeRequest(ctxSendMetric, client, c.Endpoint, ret)
					if err != nil {
						log.Printf("Http request error: %s", err.Error())
					}
				}
			}()
		}

		<-time.After(interval)
	}
}

func MakeRequest(ctx context.Context, client *http.Client, url string, body []byte) error {
	var b bytes.Buffer

	writer := gzip.NewWriter(&b)
	_, err := writer.Write(body)
	if err != nil {
		return fmt.Errorf("failed init compress writer: %v", err.Error())
	}
	writer.Close()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b.Bytes()))
	if err != nil {
		return fmt.Errorf("new request error: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("http request error: %s", err.Error())
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			return fmt.Errorf("failed init compress reader: %s", err.Error())
		}
		defer reader.Close()

		response.Body = io.NopCloser(reader)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("the request was not executed successfully")
	}

	return nil
}
