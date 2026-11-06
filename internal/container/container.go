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
	"github.com/DaleCoopTP/grishin_lab1/internal/http/middleware"
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
	closeResources func(context.Context)
	closeOnce      sync.Once
	closed         chan struct{}
}

func New(ctx context.Context, cfg config.Config) (_ *Container, err error) {
	logger := telemetry.NewLogger(cfg.LogLevel)
	sdk := telemetry.NewSDK(ctx, logger)
	var closePool func()
	closeResources := func(ctx context.Context) {
		if err := sdk.Shutdown(ctx); err != nil {
			logger.WarnContext(ctx, "telemetry shutdown failed", "error.type", fmt.Sprintf("%T", err))
		}
		if closePool != nil {
			closePool()
		}
	}
	defer func() {
		if err != nil {
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
			defer cancel()
			closeResources(closeCtx)
		}
	}()
	meter := sdk.Metrics.Meter("trip-service")
	queryTracer, err := postgres.NewQueryTracer(sdk.Traces, meter)
	if err != nil {
		return nil, err
	}
	pool, err := postgres.Open(ctx, cfg, queryTracer)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	closePool = pool.Close
	txManager := transaction.NewTxManager(pool, cfg.DatabaseQueryTimeout)
	repository := triprepository.NewRepository(txManager, cfg.DatabaseQueryTimeout)
	events, err := telemetry.NewTripEvents(meter, logger, repository)
	if err != nil {
		return nil, err
	}
	createTrip := create.New(repository, txManager, events)
	getTrip := get.New(repository)
	finishTrip := finish.New(repository, txManager, events)
	writer := response.New()
	createHandler := httpcreate.New(createTrip, writer)
	getHandler := httpget.New(getTrip, writer)
	finishHandler := httpfinish.New(finishTrip, writer)
	healthHandler := health.New(writer)
	readyHandler := ready.New(pool, cfg.DatabaseQueryTimeout, writer)
	router := httptransport.NewRouter(createHandler, getHandler, finishHandler, healthHandler, readyHandler, writer)
	handler, err := middleware.Wrap(router, writer, logger, sdk.Traces, meter)
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	return &Container{
		Server:         server,
		Logger:         logger,
		closeResources: closeResources,
		closed:         make(chan struct{}),
	}, nil
}

func (c *Container) Close(ctx context.Context) error {
	c.closeOnce.Do(func() {
		go func() {
			c.closeResources(ctx)
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
