package config

import "os"

type Config struct {
	GRPCPort string
	HTTPPort string
}

func Load() Config {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	return Config{
		GRPCPort: grpcPort,
		HTTPPort: httpPort,
	}
}

func (c Config) GRPCAddress() string {
	return ":" + c.GRPCPort
}

func (c Config) HTTPAddress() string {
	return ":" + c.HTTPPort
}
