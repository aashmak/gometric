package server

type Config struct {
	ListenAddr    string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set listen address"`
	StoreInterval int    `long:"store_interval" short:"i" env:"STORE_INTERVAL" default:"300" description:"set interval store to file"`
	StoreFile     string `long:"store_file" short:"f" env:"STORE_FILE" default:"/tmp/devops-metrics-db.json" description:"set store file"`
	Restore       bool   `long:"restore" short:"r" env:"RESTORE" description:"autorestore from file"`
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "127.0.0.1:8080",
	}
}
