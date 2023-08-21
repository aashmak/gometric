package agent

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"time"
)

type Config struct {
	ConfigFile     string `long:"config" short:"c" env:"CONFIG" default:"" description:"set config file"`
	EndpointAddr   string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set remote metric collector"`
	UseGrpc        bool   `long:"grpc" env:"GRPC" description:"use the grpc protocol to send" json:"grpc,omitempty"`
	ReportInterval int    `long:"report_interval" short:"r" env:"REPORT_INTERVAL" default:"10" description:"set report interval"`
	PollInterval   int    `long:"poll_interval" short:"p" env:"POLL_INTERVAL" default:"2" description:"set poll interval"`
	KeySign        string `long:"key" short:"k" env:"KEY" description:"set key for signing"`
	RSAPublicKey   string `long:"crypto-key" env:"CRYPTO_KEY" description:"set rsa-public-key file"`
	LogLevel       string `long:"log_level" env:"LOG_LEVEL" default:"info" description:"set log level"`
	LogFile        string `long:"log_file" env:"LOG_FILE" default:"" description:"set log file"`
	RateLimit      int    `long:"rate_limit" short:"l" env:"RATE_LIMIT" default:"1" description:"set rate limit"`
	Version        bool   `long:"version" short:"v" description:"print current version"`
}

func ParseConfigFile(cfg *Config) error {
	if cfg.ConfigFile == "" {
		return nil
	}

	cfgTmp := struct {
		Address        string `json:"address,omitempty"`
		UseGrpc        bool   `json:"grpc,omitempty"`
		ReportInterval string `json:"report_interval,omitempty"`
		PollInterval   string `json:"poll_interval,omitempty"`
		RSAPublicKey   string `json:"crypto_key,omitempty"`
	}{}

	data, err := readFile(cfg.ConfigFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &cfgTmp)
	if err != nil {
		return err
	}

	if cfg.EndpointAddr == "127.0.0.1:8080" {
		cfg.EndpointAddr = cfgTmp.Address
	}

	if !cfg.UseGrpc {
		cfg.UseGrpc = cfgTmp.UseGrpc
	}

	if cfg.ReportInterval == 10 {
		d, err := time.ParseDuration(cfgTmp.ReportInterval)
		if err != nil {
			return err
		}
		cfg.ReportInterval = int(d / time.Second)
	}

	if cfg.PollInterval == 2 {
		d, err := time.ParseDuration(cfgTmp.PollInterval)
		if err != nil {
			return err
		}
		cfg.PollInterval = int(d / time.Second)
	}

	if cfg.RSAPublicKey == "" {
		cfg.RSAPublicKey = cfgTmp.RSAPublicKey
	}

	return nil
}

func readFile(file string) ([]byte, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0777)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	data := make([]byte, 0)

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return []byte{}, err
			}
		}

		data = append(data, line...)
	}

	return data, nil
}
