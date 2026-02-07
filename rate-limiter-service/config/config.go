package config

import "os"

type Config struct {
	DB             string
	DBPort         string
	JaegerHost     string
	JaegerGRPCPort string
}

func GetConfig() Config {
	return Config{
		DB:             os.Getenv("DB_CONSUL"),
		DBPort:         os.Getenv("DBPORT_CONSUL"),
		JaegerHost:     os.Getenv("JAEGER_HOST"),
		JaegerGRPCPort: os.Getenv("JAEGER_GRPC_PORT"),
	}
}
