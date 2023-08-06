package server

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"time"
)

// Config описывает структуру с настройками сервера.
type Config struct {
	ConfigFile    string `long:"config" short:"c" env:"CONFIG" default:"" description:"set config file"`
	ListenAddr    string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set listen address"`
	TrustedSubnet string `long:"trusted_subnet" short:"t" env:"TRUSTED_SUBNET" description:"set trusted subnet (example: 10.0.0.0/8)"`
	StoreInterval int    `long:"store_interval" short:"i" env:"STORE_INTERVAL" default:"300" description:"set interval store to file"`
	StoreFile     string `long:"store_file" short:"f" env:"STORE_FILE" default:"/tmp/devops-metrics-db.json" description:"set store file"`
	Restore       bool   `long:"restore" short:"r" env:"RESTORE" description:"autorestore from file"`
	KeySign       string `long:"key" short:"k" env:"KEY" description:"set key for signing"`
	RSAPrivateKey string `long:"crypto-key" env:"CRYPTO_KEY" description:"set rsa-private-key file"`
	DatabaseDSN   string `long:"database" short:"d" env:"DATABASE_DSN" description:"set database dsn"`
	LogLevel      string `long:"log_level" env:"LOG_LEVEL" default:"info" description:"set log level"`
	LogFile       string `long:"log_file" env:"LOG_FILE" default:"" description:"set log file"`
	Version       bool   `long:"version" short:"v" description:"print current version"`
}

// DefaultConfig возвращает стандартные настройки сервера.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "127.0.0.1:8080",
		KeySign:    "",
		StoreFile:  "/tmp/devops-metrics-db.json",
	}
}

func ParseConfigFile(cfg *Config) error {
	if cfg.ConfigFile == "" {
		return nil
	}

	cfgTmp := struct {
		ListenAddr    string `json:"address,omitempty"`
		TrustedSubnet string `json:"trusted_subnet,omitempty"`
		Restore       bool   `json:"restore,omitempty"`
		StoreInterval string `json:"store_interval,omitempty"`
		StoreFile     string `json:"store_file,omitempty"`
		DatabaseDSN   string `json:"database_dsn,omitempty"`
		RSAPrivateKey string `json:"crypto_key,omitempty"`
	}{}

	data, err := readFile(cfg.ConfigFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &cfgTmp)
	if err != nil {
		return err
	}

	if cfg.ListenAddr == "127.0.0.1:8080" {
		cfg.ListenAddr = cfgTmp.ListenAddr
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = cfgTmp.TrustedSubnet
	}

	if !cfg.Restore {
		cfg.Restore = cfgTmp.Restore
	}

	if cfg.StoreInterval == 300 {
		d, err := time.ParseDuration(cfgTmp.StoreInterval)
		if err != nil {
			return err
		}
		cfg.StoreInterval = int(d / time.Second)
	}

	if cfg.StoreFile == "/tmp/devops-metrics-db.json" {
		cfg.StoreFile = cfgTmp.StoreFile
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = cfgTmp.DatabaseDSN
	}

	if cfg.RSAPrivateKey == "" {
		cfg.RSAPrivateKey = cfgTmp.RSAPrivateKey
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
