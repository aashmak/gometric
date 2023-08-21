package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"gometric/internal/crypto"
	"gometric/internal/logger"

	"gometric/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	api "gometric/internal/api"
)

type Collector struct {
	Endpoint          string
	UseGrpc           bool
	ReportIntervalSec int
	Metrics           []metrics.Metrics
	KeySign           string
	RSAPublicKey      string
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

func (c *Collector) SendMetric(ctx context.Context, wg *sync.WaitGroup) {
	var interval = time.Duration(c.ReportIntervalSec) * time.Second

	client := &http.Client{
		Timeout: interval,
	}

	var pubKey *rsa.PublicKey
	var err error
	if c.RSAPublicKey != "" {
		pubKey, err = crypto.NewPublicKey(c.RSAPublicKey)
		if err != nil {
			logger.Error("new public key error", err)
			return
		}
	}

	requestQueue := make(chan *bytes.Buffer)

	// Create worker pool
	for i := 0; i < c.RateLimit; i++ {
		workerID := i + 1
		wg.Add(1)
		go func() {
			var err error

			for request := range requestQueue {
				if c.UseGrpc {
					// send to endpoint via the grpc protocol
					err = MakeGrpcRequest(ctx, c.Endpoint, pubKey, request)
				} else {
					// send to endpoint via the http protocol
					httpHost := "http://" + c.Endpoint + "/update/"
					err = MakeRequest(ctx, client, httpHost, pubKey, request)
				}

				if err != nil {
					logger.Error(fmt.Sprintf("[Worker #%d]", workerID), err)
				} else {
					logger.Debug(fmt.Sprintf("[Worker #%d] the request was executed successfully", workerID))
				}
			}
			wg.Done()
		}()
	}

	for {
		select {
		case <-ctx.Done():
			close(requestQueue)
			logger.Debug("SendMetric() stopped")
			return

		default:
			for _, metric := range c.Metrics {
				// sign if key is not empty
				if c.KeySign == "" {
					metric.Sign(c.KeySign)
				}

				metricJSON, err := json.Marshal(metric)
				if err != nil {
					logger.Error("", err)
				} else {
					requestQueue <- bytes.NewBuffer(metricJSON)
				}
			}
		}

		<-time.After(interval)
	}
}

func MakeRequest(ctx context.Context, client *http.Client, url string, pubKey *rsa.PublicKey, body *bytes.Buffer) error {
	var b bytes.Buffer

	gzipWriter := gzip.NewWriter(&b)
	_, err := gzipWriter.Write(body.Bytes())
	if err != nil {
		return fmt.Errorf("failed init compress writer: %v", err.Error())
	}
	gzipWriter.Close()

	if pubKey != nil {
		var enc []byte
		enc, err = crypto.Encrypt(pubKey, &b)
		if err != nil {
			return fmt.Errorf("encrypt failed: %v", err.Error())
		}

		b = *bytes.NewBuffer(enc)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b.Bytes()))
	if err != nil {
		return fmt.Errorf("new request error: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Add("X-Real-IP", "10.0.0.10")

	if pubKey != nil {
		request.Header.Add("Content-Encrypt", "rsa")
	}

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

func MakeGrpcRequest(ctx context.Context, host string, pubKey *rsa.PublicKey, body *bytes.Buffer) error {
	var b bytes.Buffer

	gzipWriter := gzip.NewWriter(&b)
	_, err := gzipWriter.Write(body.Bytes())
	if err != nil {
		return fmt.Errorf("failed init compress writer: %v", err.Error())
	}
	gzipWriter.Close()

	if pubKey != nil {
		var enc []byte
		enc, err = crypto.Encrypt(pubKey, &b)
		if err != nil {
			return fmt.Errorf("encrypt failed: %v", err.Error())
		}

		b = *bytes.NewBuffer(enc)
	}

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	grpcClient := api.NewGometricAPIClient(conn)

	md := metadata.New(map[string]string{
		"content-encoding": "gzip",
		"x-real-ip":        "10.0.0.10",
	})

	if pubKey != nil {
		md.Append("content-encrypt", "rsa")
	}

	ctx2 := metadata.NewOutgoingContext(ctx, md)

	_, err = grpcClient.UpdateMeticValue(ctx2, &api.Request{
		Bytes: b.Bytes(),
	})

	return err
}
