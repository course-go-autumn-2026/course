package container

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/DaleCoopTP/grishin_lab1/internal/config"
)

func TestCloseWaitsForResources(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	app := &Container{
		closed: make(chan struct{}),
		closeResources: func(context.Context) {
			close(started)
			<-release
		},
	}
	done := make(chan error, 1)
	go func() { done <- app.Close(context.Background()) }()
	<-started
	select {
	case err := <-done:
		t.Fatalf("Close returned before resource cleanup finished: %v", err)
	default:
	}
	releaseOnce.Do(func() { close(release) })
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestCloseCancellationDoesNotAbandonCleanup(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	var calls atomic.Int32
	app := &Container{
		closed: make(chan struct{}),
		closeResources: func(context.Context) {
			calls.Add(1)
			<-release
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := app.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	releaseOnce.Do(func() { close(release) })
	if err := app.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("cleanup calls: got %d, want 1", got)
	}
}

func TestConcurrentCloseReleasesResourcesOnce(t *testing.T) {
	var calls atomic.Int32
	app := &Container{
		closed:         make(chan struct{}),
		closeResources: func(context.Context) { calls.Add(1) },
	}
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := app.Close(context.Background()); err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	if got := calls.Load(); got != 1 {
		t.Fatalf("cleanup calls: got %d, want 1", got)
	}
}

func TestNewRejectsInvalidDatabaseConfiguration(t *testing.T) {
	app, err := New(context.Background(), config.Config{DatabaseURL: "://"})
	if err == nil || app != nil {
		t.Fatalf("expected initialization failure without an application, got %v, %v", app, err)
	}
}
