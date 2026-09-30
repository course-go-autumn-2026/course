package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr              string
	HTTPReadTimeout       time.Duration
	HTTPReadHeaderTimeout time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	LogLevel              string
	ShutdownTimeout       time.Duration
	Database              DatabaseConfig
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
}

func Load() (Config, error) {
	httpAddr, err := requiredString("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	httpReadTimeout, err := requiredDuration("HTTP_READ_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	httpReadHeaderTimeout, err := requiredDuration("HTTP_READ_HEADER_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	httpWriteTimeout, err := requiredDuration("HTTP_WRITE_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	httpIdleTimeout, err := requiredDuration("HTTP_IDLE_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	logLevel, err := requiredString("LOG_LEVEL")
	if err != nil {
		return Config{}, err
	}
	logLevel = strings.ToLower(logLevel)
	if _, ok := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}[logLevel]; !ok {
		return Config{}, fmt.Errorf("LOG_LEVEL: unsupported value %q", logLevel)
	}
	shutdownTimeout, err := requiredDuration("SHUTDOWN_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	dbURL, err := requiredString("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	maxConns, err := requiredInt32("DATABASE_MAX_CONNS")
	if err != nil {
		return Config{}, err
	}
	minConns, err := requiredInt32("DATABASE_MIN_CONNS")
	if err != nil {
		return Config{}, err
	}
	maxLifetime, err := requiredDuration("DATABASE_MAX_CONN_LIFETIME")
	if err != nil {
		return Config{}, err
	}
	connectTimeout, err := requiredDuration("DATABASE_CONNECT_TIMEOUT")
	if err != nil {
		return Config{}, err
	}
	queryTimeout, err := requiredDuration("DATABASE_QUERY_TIMEOUT")
	if err != nil {
		return Config{}, err
	}

	if maxConns < 1 {
		return Config{}, fmt.Errorf("DATABASE_MAX_CONNS: must be greater than zero")
	}
	if minConns < 0 {
		return Config{}, fmt.Errorf("DATABASE_MIN_CONNS: must not be negative")
	}
	if minConns > maxConns {
		return Config{}, fmt.Errorf("DATABASE_MIN_CONNS: must be <= DATABASE_MAX_CONNS")
	}
	if shutdownTimeout <= 0 || httpReadTimeout <= 0 || httpReadHeaderTimeout <= 0 || httpWriteTimeout <= 0 || httpIdleTimeout <= 0 || maxLifetime <= 0 || connectTimeout <= 0 || queryTimeout <= 0 {
		return Config{}, fmt.Errorf("timeouts and connection lifetime must be greater than zero")
	}

	return Config{
		HTTPAddr:              httpAddr,
		HTTPReadTimeout:       httpReadTimeout,
		HTTPReadHeaderTimeout: httpReadHeaderTimeout,
		HTTPWriteTimeout:      httpWriteTimeout,
		HTTPIdleTimeout:       httpIdleTimeout,
		LogLevel:              logLevel,
		ShutdownTimeout:       shutdownTimeout,
		Database: DatabaseConfig{
			URL:             dbURL,
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: maxLifetime,
			ConnectTimeout:  connectTimeout,
			QueryTimeout:    queryTimeout,
		},
	}, nil
}

func requiredString(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func requiredDuration(name string) (time.Duration, error) {
	value, err := requiredString(name)
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration %q: %w", name, value, err)
	}
	return duration, nil
}

func requiredInt32(name string) (int32, error) {
	value, err := requiredString(name)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q: %w", name, value, err)
	}
	return int32(parsed), nil
}
