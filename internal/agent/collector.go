package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"gometric/internal/logger"
	"io"
	"net/http"
	"time"

	"gometric/internal/metrics"
)

type Collector struct {
	Endpoint          string
	ReportIntervalSec int
	Metrics           []metrics.Metrics
	KeySign           []byte
	RateLimit         int
}

func (c *Collector) RegisterMetric(name string, value interface{}) error {
	if c.Metrics == nil {
		c.Metrics = make([]metrics.Metrics, 0)
	}

	for _, v := range c.Metrics {
		if v.ID == name {
			return fmt.Errorf("metric %s already exists", name)
		}
	}

	tmp := metrics.Metrics{
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

	logger.Debug(fmt.Sprintf("metric %s registered successfully", name))
	return nil
}

func (c *Collector) SendMetric(ctx context.Context) {
	var interval = time.Duration(c.ReportIntervalSec) * time.Second

	client := &http.Client{
		Timeout: interval,
	}

	requestQueue := make(chan []byte, 50)

	//Create worker pool
	for i := 0; i < c.RateLimit; i++ {
		workerID := i + 1
		go func(workerID int, ctx context.Context, client *http.Client, url string, requestQueue <-chan []byte) {
			for req := range requestQueue {
				err := MakeRequest(ctx, client, c.Endpoint, req)
				if err != nil {
					logger.Error(fmt.Sprintf("[Worker #%d]", workerID), err)
				} else {
					logger.Debug(fmt.Sprintf("[Worker #%d] the request was executed successfully", workerID))
				}
			}
		}(workerID, ctx, client, c.Endpoint, requestQueue)
	}

	for {
		select {
		case <-ctx.Done():
			close(requestQueue)
			logger.Info("SendMetric stopped")
			return

		default:
			for _, metric := range c.Metrics {
				//sign if key is not empty
				if !bytes.Equal(c.KeySign, []byte{}) {
					metric.Sign(c.KeySign)
				}

				metricJSON, err := json.Marshal(metric)
				if err != nil {
					logger.Error("", err)
				} else {
					requestQueue <- metricJSON
				}
			}
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
