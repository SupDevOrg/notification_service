package config

import "os"

type Config struct {
	GRPCPort string
}

func Load() Config {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "9090"
	}

	return Config{
		GRPCPort: port,
	}
}

func (c Config) GRPCAddress() string {
	return ":" + c.GRPCPort
}
