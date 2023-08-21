package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"fmt"
	"gometric/internal/crypto"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	api "gometric/internal/api"
)

func GrpcRequest(ctx context.Context, method string, clientIP string, pubKey *rsa.PublicKey, body *bytes.Buffer) (*empty.Empty, *api.Response, error) {
	var b bytes.Buffer

	gzipWriter := gzip.NewWriter(&b)
	_, err := gzipWriter.Write(body.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("failed init compress writer: %v", err.Error())
	}
	gzipWriter.Close()

	if pubKey != nil {
		var enc []byte
		enc, err = crypto.Encrypt(pubKey, &b)
		if err != nil {
			return nil, nil, fmt.Errorf("encrypt failed: %v", err.Error())
		}

		b = *bytes.NewBuffer(enc)
	}

	conn, err := grpc.Dial(":50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	client1 := api.NewGometricAPIClient(conn)

	md := metadata.New(map[string]string{
		"content-encoding": "gzip",
	})

	if clientIP != "" {
		md.Append("x-real-ip", clientIP)
	}

	if pubKey != nil {
		md.Append("content-encrypt", "rsa")
	}

	ctx2 := metadata.NewOutgoingContext(ctx, md)

	switch method {
	case "UpdateMeticValue":
		_, err = client1.UpdateMeticValue(ctx2, &api.Request{
			Bytes: b.Bytes(),
		})

		return nil, nil, err

	case "GetMeticValue":
		var resp *api.Response
		resp, err = client1.GetMeticValue(ctx2, &api.Request{
			Bytes: b.Bytes(),
		})

		return nil, resp, err
	}

	return nil, nil, nil
}

func TestGrpcServer1(t *testing.T) {
	pubkeyData := `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA0awwJk7i4hcTWfIrqDJN
jieE+CSusRfAmKF/hYxT8sEoBrP7ZG3aAw7qPY58nfPwXzJarsp+3Prf0OobthPP
l2EOihTX6FDj9m3WhmzHMk4Ryf2OOsdAo9N4vcmWC0LW2GuiooVslqTVIvlyk3VL
xSpclBV3Uer1irJdMJAP7+QedMKdx5x67twqnZ7r9mFrug3q6DxA9E8B1Su9OrAN
5EpNpXHZVH9acxgbDu8jVs9SW1NAPvOapcQ/hzC4VjzfGAVfkBU3CAqwWdlzrusA
G3o9T8q7GAj5rkUPnsvQxKZifnmbnbOzozp4iODYw5eLfLuOS5KwbADNmrX6Fb1k
rdGLTLWXYNie/nnIn8zzHeohXfiwoHfoljZvMFsPUiWctjOnIQEh6qCiOFeOOHJm
nYFcQ9oEybWhiZ9hetl6Pr1aHz0yL5TeGoI4/GPynKBn29V1LIZ2g81/yQfnQUXH
K6y84KAq013Uz39lM8KQhpsB8btoMJqlKBzGuMAT64HZanzP39SKvLJADnWigQV3
kPQJ3Cxmn9hHqGM/gi7zAo4gQ1mUAbo2Fbfm/IdjDfIrJfLzedUiYYEzwVV2kWrf
Sd+vi3l4X9RGKlbLZHYCFXciLG+jJm7aa5EogwAFIGDNNAqSAOX7djJIXT//fQmF
wxivQLf/ly6Inr2A9FDIJLECAwEAAQ==
-----END PUBLIC KEY-----
`

	privkeyData := `-----BEGIN PRIVATE KEY-----
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
YhlnDV2uwRzXWt/4nCeyXdVQh+TUkm8=
-----END PRIVATE KEY-----
`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	cfg.ListenAddrGrpc = "127.0.0.1:50052"
	cfg.TrustedSubnet = "10.0.0.0/24"

	serv := NewServer(ctx, cfg)
	serv.RSAPrivateKey, _ = crypto.NewPrivateKeyFromBytes([]byte(privkeyData))
	go serv.GrpcListenAndServe(cfg.ListenAddrGrpc)

	// init client pub rsa-key
	var pubKey *rsa.PublicKey
	pubKey, _ = crypto.NewPublicKeyFromBytes([]byte(pubkeyData))

	tests := []struct {
		name               string
		action             string
		requestBody        []byte
		responseStatusCode codes.Code
		responseBody       string
		clientIP           string
	}{
		{
			name:               "update gauge #1",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			responseStatusCode: codes.OK,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update gauge #2",
			action:             "update",
			requestBody:        []byte(`{"id":"BuckHashSys","type":"gauge","value":3877}`),
			responseStatusCode: codes.OK,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update counter #3",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","delta":1}`),
			responseStatusCode: codes.OK,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update non gauge value #4",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","delta":1}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update non counter value #5",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"counter","value":2}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update unsupport type #6",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount","type":"integer","value":2}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update empty body #7",
			action:             "update",
			requestBody:        []byte(`{}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update empty body #8",
			action:             "update",
			requestBody:        []byte(`{"id":"PollCount"}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "get value gauge #1",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			responseStatusCode: codes.OK,
			responseBody:       `{"id":"Alloc","type":"gauge","value":1907608}`,
			clientIP:           "10.0.0.10",
		},
		{
			name:               "get value counter #2",
			action:             "value",
			requestBody:        []byte(`{"id":"PollCount","type":"counter"}`),
			responseStatusCode: codes.OK,
			responseBody:       `{"id":"PollCount","type":"counter","delta":1}`,
			clientIP:           "10.0.0.10",
		},
		{
			name:               "get unknown value #3",
			action:             "value",
			requestBody:        []byte(`{"id":"New","type":"counter"}`),
			responseStatusCode: codes.NotFound,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "get unknown value #4",
			action:             "value",
			requestBody:        []byte(`{}`),
			responseStatusCode: codes.NotFound,
			responseBody:       "",
			clientIP:           "10.0.0.10",
		},
		{
			name:               "update gauge from untrusted IP",
			action:             "update",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge","value":1907608}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       "",
			clientIP:           "10.10.10.10",
		},
		{
			name:               "get value from untrusted IP",
			action:             "value",
			requestBody:        []byte(`{"id":"Alloc","type":"gauge"}`),
			responseStatusCode: codes.PermissionDenied,
			responseBody:       `{"id":"Alloc","type":"gauge","value":1907608}`,
			clientIP:           "10.0.0.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "update" {
				_, _, err := GrpcRequest(ctx, "UpdateMeticValue", tt.clientIP, pubKey, bytes.NewBuffer(tt.requestBody))
				if err != nil {
					if e, ok := status.FromError(err); ok {
						if tt.responseStatusCode != e.Code() {
							t.Errorf("Error")
						}
					}
				}
			}

			if tt.action == "value" {
				_, resp, err := GrpcRequest(ctx, "GetMeticValue", tt.clientIP, pubKey, bytes.NewBuffer(tt.requestBody))
				if err != nil {
					if e, ok := status.FromError(err); ok {
						if tt.responseStatusCode != e.Code() || tt.responseBody != string(resp.GetBytes()) {
							t.Errorf("Error")
						}
					}
				} else {
					if tt.responseBody != string(resp.GetBytes()) {
						t.Errorf("Error")
					}
				}
			}
		})
	}
}
