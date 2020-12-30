package config

import (
	"os"
	"time"
)

// TODO: more robust config
// TODO: more robust waiting for DB up

type DatabaseConfig struct {
	Host           string
	Username       string
	Password       string
	DatabaseName   string
	Port           string
	Timezone       string
	ConnectRetries int
	ConnectBackoff time.Duration
}

type HTTPConfig struct {
	Address string
}

type GrpcConfig struct {
	Address string
}

func (conf *GrpcConfig) Load() {
	conf.Address = GetEnv("GRPC_ADDRESS", conf.Address)
}

func (conf *HTTPConfig) Load() {
	conf.Address = GetEnv("HTTP_ADDRESS", conf.Address)
	// TODO: expose timeouts in conf
}

func (conf *DatabaseConfig) Load() {
	// env vars take precedence over existing conf setting
	conf.Host = GetEnv("DB_HOST", conf.Host)
	conf.Username = GetEnv("DB_USERNAME", conf.Username)
	conf.Password = GetEnv("DB_PASSWORD", conf.Password)
	conf.DatabaseName = GetEnv("DB_NAME", conf.DatabaseName)
	conf.Port = GetEnv("DB_PORT", conf.Port)
	conf.Timezone = GetEnv("DB_TIMEZONE", conf.Timezone)
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
