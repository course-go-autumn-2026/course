//go:build integration

package trip_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	domaintrip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/infra/postgres/transaction"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
	repository "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/repository/trip"
)

func TestRepositorySaveAndGet(t *testing.T) {
	pool := integrationPool(t)
	repo := repository.New(pool, 3*time.Second)
	value := integrationTrip(t)
	t.Cleanup(func() {
		cleanupTrip(t, pool, value.ID)
	})

	if err := repo.Save(context.Background(), value); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	actual, err := repo.GetByID(context.Background(), value.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if actual.ID != value.ID || actual.Status != domaintrip.StatusActive || actual.Price != value.Price {
		t.Errorf("GetByID() = %#v, want saved trip %#v", actual, value)
	}
}

func TestRepositoryLocksAndUpdatesTrip(t *testing.T) {
	pool := integrationPool(t)
	repo := repository.New(pool, 3*time.Second)
	value := integrationTrip(t)
	t.Cleanup(func() {
		cleanupTrip(t, pool, value.ID)
	})
	if err := repo.Save(context.Background(), value); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	locked, err := repo.GetForUpdate(context.Background(), value.ID)
	if err != nil {
		t.Fatalf("GetForUpdate() error = %v", err)
	}
	if err := locked.Finish(time.Now().UTC()); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}
	if err := repo.Update(context.Background(), locked); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	actual, err := repo.GetByID(context.Background(), value.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if actual.Status != domaintrip.StatusCompleted || actual.FinishedAt == nil {
		t.Errorf("updated trip = %#v, want completed trip", actual)
	}
}

func TestTransactionRollbackRemovesTripAndHistory(t *testing.T) {
	pool := integrationPool(t)
	repo := repository.New(pool, 3*time.Second)
	manager := transaction.New(pool)
	value := integrationTrip(t)
	failure := errors.New("rollback")
	t.Cleanup(func() {
		cleanupTrip(t, pool, value.ID)
	})

	err := manager.Do(context.Background(), func(ctx context.Context) error {
		if err := repo.Save(ctx, value); err != nil {
			return err
		}
		if err := repo.Add(ctx, model.StatusChange{TripID: value.ID, ToStatus: string(domaintrip.StatusActive), Reason: "trip created"}); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("Do() error = %v, want %v", err, failure)
	}

	var trips, history int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM trips WHERE id = $1", value.ID).Scan(&trips); err != nil {
		t.Fatalf("count trips: %v", err)
	}
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", value.ID).Scan(&history); err != nil {
		t.Fatalf("count history: %v", err)
	}
	if trips != 0 || history != 0 {
		t.Errorf("rows after rollback: trips=%d history=%d, want 0 and 0", trips, history)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required; run tripgoctl environment start first")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("open PostgreSQL pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func integrationTrip(t *testing.T) domaintrip.Trip {
	t.Helper()
	value, err := domaintrip.New(uuid.New(), model.CreateTripInput{
		UserID:     uuid.New(),
		DriverID:   uuid.New(),
		StartPoint: model.Coordinates{Latitude: 59.9398, Longitude: 30.3146},
		EndPoint:   model.Coordinates{Latitude: 59.929, Longitude: 30.3626},
		Price:      1450,
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("create test trip: %v", err)
	}
	return value
}

func cleanupTrip(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), "DELETE FROM trip_status_history WHERE trip_id = $1", id); err != nil {
		t.Errorf("delete history: %v", err)
	}
	if _, err := pool.Exec(context.Background(), "DELETE FROM trips WHERE id = $1", id); err != nil {
		t.Errorf("delete trip: %v", err)
	}
}
