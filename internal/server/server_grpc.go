package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	api "gometric/internal/api"
	"gometric/internal/crypto"
	"gometric/internal/logger"
	"gometric/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/emptypb"
)

// GetMeticValue извлекает метрики из key-value бэкенда и отсылает в формате protobuf.
// Функция также подписывает сообщение перед отправкой с помощью функции Sign().
func (s Server) GetMeticValue(ctx context.Context, request *api.Request) (*api.Response, error) {
	data := request.GetBytes()

	var metric metrics.Metrics
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("unmarshall succefull: %v", metric))

	v, err := s.Storage.Get(metric.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "")
	}

	switch metric.MType {
	case "gauge":
		if gaugeType(v) {
			v1 := v.(float64)
			metric.Value = (*float64)(&v1)

			// sign if key is not empty
			if s.KeySign != "" {
				metric.Sign(s.KeySign)
			}

			ret, err := json.Marshal(metric)
			if err != nil {
				return nil, err
			}
			logger.Debug(fmt.Sprintf("marshall succefull: %s", ret))

			resp := &api.Response{
				Bytes: ret,
			}

			return resp, status.Errorf(codes.OK, "OK")
		}
	case "counter":
		if counterType(v) {
			v1 := v.(int64)
			metric.Delta = (*int64)(&v1)

			// sign if key is not empty
			if s.KeySign != "" {
				metric.Sign(s.KeySign)
			}

			ret, err := json.Marshal(metric)
			if err != nil {
				return nil, err
			}
			logger.Debug(fmt.Sprintf("marshall succefull: %s", ret))

			resp := &api.Response{
				Bytes: ret,
			}

			return resp, status.Errorf(codes.OK, "OK")
		}
	}

	return nil, status.Errorf(codes.NotFound, "NotFound")
}

// UpdateMeticValue принимает метрики в формате protobuf и сохраняет в key-value бэкенд.
// Функция также проверяет подпись с помощью ValidMAC().
func (s Server) UpdateMeticValue(ctx context.Context, request *api.Request) (*emptypb.Empty, error) {
	data := request.GetBytes()

	var metric metrics.Metrics
	if err := json.Unmarshal(data, &metric); err != nil {
		return &emptypb.Empty{}, err
	}

	// ValidMAC if key is not epmty
	if s.KeySign != "" {
		if !metric.ValidMAC(s.KeySign) {
			logger.Debug("invalid HMAC of the data")
			return &emptypb.Empty{}, status.Errorf(codes.PermissionDenied, "invalid HMAC of the data")
		}
	}

	switch metric.MType {
	case "gauge":
		if metric.ID != "" && metric.Value != nil {
			err := s.Storage.Set(metric.ID, float64(*metric.Value))
			if err == nil {
				return &emptypb.Empty{}, status.Errorf(codes.OK, "OK")
			}
		}
	case "counter":
		// get previous counter value
		prevCounter, err := s.Storage.Get(metric.ID)
		if err != nil {
			prevCounter = int64(0)
		}

		if metric.ID != "" && metric.Delta != nil {
			err = s.Storage.Set(metric.ID, (*metric.Delta + prevCounter.(int64)))
			if err == nil {
				return &emptypb.Empty{}, status.Errorf(codes.OK, "OK")
			}
		}
	}

	return &emptypb.Empty{}, status.Errorf(codes.PermissionDenied, "")
}

// unaryTrustedSubnetInterceptor проверяет, что переданный в заголовке запроса X-Real-IP IP-адрес агента
// входит в доверенную подсеть, в противном случае возвращается код ответа codes.PermissionDenied.
func (s Server) unaryTrustedSubnetInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("x-real-ip")

		if s.TrustedSubnet != nil && len(values) > 0 {
			IPAddr := net.ParseIP(values[0])

			if IPAddr == nil || !s.TrustedSubnet.Contains(IPAddr) {
				logger.Debug(fmt.Sprintf("client IP %s is not allowed", IPAddr))
				return nil, status.Errorf(codes.PermissionDenied, fmt.Sprintf("client IP %s is not allowed", IPAddr))
			}
		}
	}

	return handler(ctx, req)
}

// unaryDecryptRSAInterceptor используется для расшифровки тела запроса с помощью RSA private-key.
func (s Server) unaryDecryptRSAInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("content-encrypt")

		if len(values) > 0 && values[0] == "rsa" {
			encryptData := bytes.NewBuffer(req.(*api.Request).GetBytes())
			decryptData, err := crypto.Decrypt(s.RSAPrivateKey, encryptData)
			if err != nil {
				return nil, err
			}
			logger.Debug("request decrypted successfully")

			req.(*api.Request).Bytes = decryptData
		}
	}

	return handler(ctx, req)
}

// unaryUnzipInterceptor используется для распаковки сжатого с помощью gzip тела сообщения.
func unaryUnzipInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("content-encoding")

		if len(values) > 0 && values[0] == "gzip" {
			gzipData := bytes.NewBuffer(req.(*api.Request).GetBytes())
			reader, err := gzip.NewReader(gzipData)
			if err != nil {
				return nil, err
			}
			defer reader.Close()

			req.(*api.Request).Bytes, _ = io.ReadAll(reader)
		}
	}

	return handler(ctx, req)
}

func (s Server) GrpcListenAndServe(addr string) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.unaryTrustedSubnetInterceptor,
			s.unaryDecryptRSAInterceptor,
			unaryUnzipInterceptor,
		),
	)

	api.RegisterGometricAPIServer(grpcServer, s)

	logger.Info("Start gRPC server")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
