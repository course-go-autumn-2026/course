package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/TheJCN/Avito-Course/internal/handler"
	"github.com/TheJCN/Avito-Course/internal/repository"
	"github.com/TheJCN/Avito-Course/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	address               string
	databaseURL           string
	maxConnections        int32
	minConnections        int32
	connectTimeout        time.Duration
	queryTimeout          time.Duration
	maxConnectionLifetime time.Duration
	shutdownTimeout       time.Duration
	readTimeout           time.Duration
	readHeaderTimeout     time.Duration
	writeTimeout          time.Duration
	idleTimeout           time.Duration
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	poolConfig.MaxConns = cfg.maxConnections
	poolConfig.MinConns = cfg.minConnections
	poolConfig.MaxConnLifetime = cfg.maxConnectionLifetime
	poolConfig.ConnConfig.ConnectTimeout = cfg.connectTimeout

	ctx, cancel := context.WithTimeout(context.Background(), cfg.connectTimeout)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		cancel()
		return fmt.Errorf("open database: %w", err)
	}
	err = pool.Ping(ctx)
	cancel()
	if err != nil {
		pool.Close()
		return fmt.Errorf("ping database: %w", err)
	}
	defer pool.Close()

	trips := repository.NewTripRepository(pool, cfg.queryTimeout)
	transactions := repository.NewTxManager(pool, cfg.queryTimeout)
	tripsService := service.NewTripService(trips, transactions)
	server := &http.Server{
		Addr: cfg.address, Handler: handler.New(tripsService, pool, cfg.queryTimeout),
		ReadTimeout: cfg.readTimeout, ReadHeaderTimeout: cfg.readHeaderTimeout,
		WriteTimeout: cfg.writeTimeout, IdleTimeout: cfg.idleTimeout,
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan struct{})
	stopShutdown := context.AfterFunc(signalCtx, func() {
		defer close(done)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown", "error", err)
			if err := server.Close(); err != nil {
				slog.Error("close server", "error", err)
			}
		}
	})

	slog.Info("listening", "address", cfg.address)
	err = server.ListenAndServe()
	if !stopShutdown() {
		<-done
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}

func readConfig() (config, error) {
	var cfg config
	var err error
	if cfg.address, err = required("HTTP_ADDR"); err != nil {
		return cfg, err
	}
	if cfg.databaseURL, err = required("DATABASE_URL"); err != nil {
		return cfg, err
	}
	level, err := required("LOG_LEVEL")
	if err != nil {
		return cfg, err
	}
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return cfg, fmt.Errorf("invalid LOG_LEVEL: %w", err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
	if cfg.maxConnections, err = integer("DATABASE_MAX_CONNS"); err != nil {
		return cfg, err
	}
	if cfg.minConnections, err = integer("DATABASE_MIN_CONNS"); err != nil {
		return cfg, err
	}
	if cfg.maxConnections < 1 || cfg.minConnections < 0 || cfg.minConnections > cfg.maxConnections {
		return cfg, errors.New("invalid database pool size")
	}
	if cfg.connectTimeout, err = duration("DATABASE_CONNECT_TIMEOUT"); err != nil {
		return cfg, err
	}
	if cfg.queryTimeout, err = duration("DATABASE_QUERY_TIMEOUT"); err != nil {
		return cfg, err
	}
	if cfg.maxConnectionLifetime, err = duration("DATABASE_MAX_CONN_LIFETIME"); err != nil {
		return cfg, err
	}
	if cfg.shutdownTimeout, err = duration("SHUTDOWN_TIMEOUT"); err != nil {
		return cfg, err
	}
	if cfg.readTimeout, err = duration("HTTP_READ_TIMEOUT"); err != nil {
		return cfg, err
	}
	if cfg.readHeaderTimeout, err = duration("HTTP_READ_HEADER_TIMEOUT"); err != nil {
		return cfg, err
	}
	if cfg.writeTimeout, err = duration("HTTP_WRITE_TIMEOUT"); err != nil {
		return cfg, err
	}
	if cfg.idleTimeout, err = duration("HTTP_IDLE_TIMEOUT"); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func required(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func integer(name string) (int32, error) {
	value, err := required(name)
	if err != nil {
		return 0, err
	}
	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return int32(number), nil
}

func duration(name string) (time.Duration, error) {
	value, err := required(name)
	if err != nil {
		return 0, err
	}
	result, err := time.ParseDuration(value)
	if err != nil || result <= 0 {
		return 0, fmt.Errorf("invalid %s: %q", name, value)
	}
	return result, nil
}
