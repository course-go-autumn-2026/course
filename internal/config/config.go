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
	var cfg Config
	var err error

	if cfg.HTTPAddr, err = required("HTTP_ADDR"); err != nil {
		return Config{}, err
	}

	if cfg.HTTPReadTimeout, err = positiveDuration("HTTP_READ_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if cfg.HTTPReadHeaderTimeout, err = positiveDuration("HTTP_READ_HEADER_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if cfg.HTTPWriteTimeout, err = positiveDuration("HTTP_WRITE_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if cfg.HTTPIdleTimeout, err = positiveDuration("HTTP_IDLE_TIMEOUT"); err != nil {
		return Config{}, err
	}
	if cfg.LogLevel, err = required("LOG_LEVEL"); err != nil {
		return Config{}, err
	}

	if cfg.ShutdownTimeout, err = positiveDuration("SHUTDOWN_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if cfg.Database.URL, err = required("DATABASE_URL"); err != nil {
		return Config{}, err
	}

	if cfg.Database.MaxConns, err = positiveInt32("DATABASE_MAX_CONNS"); err != nil {
		return Config{}, err
	}

	if cfg.Database.MinConns, err = nonNegativeInt32("DATABASE_MIN_CONNS"); err != nil {
		return Config{}, err
	}

	if cfg.Database.MinConns > cfg.Database.MaxConns {
		return Config{}, fmt.Errorf("DATABASE_MIN_CONNS cannot exceed DATABASE_MAX_CONNS")
	}

	if cfg.Database.MaxConnLifetime, err = positiveDuration("DATABASE_MAX_CONN_LIFETIME"); err != nil {
		return Config{}, err
	}

	if cfg.Database.ConnectTimeout, err = positiveDuration("DATABASE_CONNECT_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if cfg.Database.QueryTimeout, err = positiveDuration("DATABASE_QUERY_TIMEOUT"); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func required(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}

	return value, nil
}

func positiveDuration(key string) (time.Duration, error) {
	value, err := required(key)
	if err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return duration, nil
}

func positiveInt32(key string) (int32, error) {
	value, err := nonNegativeInt32(key)
	if err != nil {
		return 0, err
	}

	if value == 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return value, nil
}

func nonNegativeInt32(key string) (int32, error) {
	value, err := required(key)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	if number < 0 {
		return 0, fmt.Errorf("%s cannot be negative", key)
	}

	return int32(number), nil
}
