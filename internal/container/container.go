package container

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/DaleCoopTP/grishin_lab1/internal/config"
	httptransport "github.com/DaleCoopTP/grishin_lab1/internal/http"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/health"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/ready"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
	httpcreate "github.com/DaleCoopTP/grishin_lab1/internal/http/trip/create"
	httpfinish "github.com/DaleCoopTP/grishin_lab1/internal/http/trip/finish"
	httpget "github.com/DaleCoopTP/grishin_lab1/internal/http/trip/get"
	"github.com/DaleCoopTP/grishin_lab1/internal/infra/postgres"
	"github.com/DaleCoopTP/grishin_lab1/internal/infra/postgres/transaction"
	"github.com/DaleCoopTP/grishin_lab1/internal/infra/telemetry"
	triprepository "github.com/DaleCoopTP/grishin_lab1/internal/repository/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/create"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/finish"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/get"
)

type Container struct {
	Server         *http.Server
	Logger         *slog.Logger
	closeResources func()
	closeOnce      sync.Once
	closed         chan struct{}
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	logger := telemetry.NewLogger(cfg.LogLevel)
	pool, err := postgres.Open(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	txManager := transaction.NewTxManager(pool, cfg.DatabaseQueryTimeout)
	repository := triprepository.NewRepository(txManager, cfg.DatabaseQueryTimeout)
	createTrip := create.New(repository, txManager)
	getTrip := get.New(repository)
	finishTrip := finish.New(repository, txManager)
	writer := response.New(logger)
	createHandler := httpcreate.New(createTrip, writer)
	getHandler := httpget.New(getTrip, writer)
	finishHandler := httpfinish.New(finishTrip, writer)
	healthHandler := health.New(writer)
	readyHandler := ready.New(pool, cfg.DatabaseQueryTimeout, writer)
	router := httptransport.NewRouter(createHandler, getHandler, finishHandler, healthHandler, readyHandler, writer)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	return &Container{
		Server:         server,
		Logger:         logger,
		closeResources: pool.Close,
		closed:         make(chan struct{}),
	}, nil
}

func (c *Container) Close(ctx context.Context) error {
	c.closeOnce.Do(func() {
		go func() {
			c.closeResources()
			close(c.closed)
		}()
	})
	select {
	case <-c.closed:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("close application resources: %w", ctx.Err())
	}
}
