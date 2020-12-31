// Package config provides application configuration implementations.
package config

import (
	"os"
	"time"
)

// TODO: more robust config
// TODO: more robust waiting for DB up

// DatabaseConfig defines the database specific configuration.
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

// HTTPConfig defines configuration specific to the REST API HTTP server.
type HTTPConfig struct {
	Address string
}

// GrpcConfig defines configuration for the GRPC server.
type GrpcConfig struct {
	Address string
}

// Load loads the GrpcConfig options from env vars overriding existing values.
func (conf *GrpcConfig) Load() {
	conf.Address = GetEnv("GRPC_ADDRESS", conf.Address)
}

// Load loads the HTTPConfig options from env vars overriding existing values.
func (conf *HTTPConfig) Load() {
	conf.Address = GetEnv("HTTP_ADDRESS", conf.Address)
	// TODO: expose timeouts in conf
}

// Load loads the DatabaseConfig options from env vars overriding existing values.
func (conf *DatabaseConfig) Load() {
	// env vars take precedence over existing conf setting
	conf.Host = GetEnv("DB_HOST", conf.Host)
	conf.Username = GetEnv("DB_USERNAME", conf.Username)
	conf.Password = GetEnv("DB_PASSWORD", conf.Password)
	conf.DatabaseName = GetEnv("DB_NAME", conf.DatabaseName)
	conf.Port = GetEnv("DB_PORT", conf.Port)
	conf.Timezone = GetEnv("DB_TIMEZONE", conf.Timezone)
}

// GetEnv gets the said env variable returning the defaultValue if not set.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
