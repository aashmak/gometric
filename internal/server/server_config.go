package server

// Config описывает структуру с настройками сервера.
type Config struct {
	ListenAddr    string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set listen address"`
	StoreInterval int    `long:"store_interval" short:"i" env:"STORE_INTERVAL" default:"300" description:"set interval store to file"`
	StoreFile     string `long:"store_file" short:"f" env:"STORE_FILE" default:"/tmp/devops-metrics-db.json" description:"set store file"`
	Restore       bool   `long:"restore" short:"r" env:"RESTORE" description:"autorestore from file"`
	KeySign       string `long:"key" short:"k" env:"KEY" description:"set key for signing"`
	DatabaseDSN   string `long:"database" short:"d" env:"DATABASE_DSN" description:"set database dsn"`
	LogLevel      string `long:"log_level" env:"LOG_LEVEL" default:"info" description:"set log level"`
	LogFile       string `long:"log_file" env:"LOG_FILE" default:"" description:"set log file"`
}

// DefaultConfig возвращает стандартные настройки сервера.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "127.0.0.1:8080",
		KeySign:    "",
		StoreFile:  "/tmp/devops-metrics-db.json",
	}
}
