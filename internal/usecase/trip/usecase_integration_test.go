package trip_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/infra/postgres/transaction"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
	triprepository "github.com/DaleCoopTP/grishin_lab1/internal/repository/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/create"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/finish"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/get"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUsecasesConcurrencyAndHistory(t *testing.T) {
	usecases, pool := testUsecases(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID, driverID := uuid.New(), uuid.New()
	cleanupDriver(t, pool, driverID)
	input := model.CreateTripInput{
		UserID: userID, DriverID: driverID,
		StartPoint: tripdomain.Point{Latitude: 59.9398, Longitude: 30.3146},
		EndPoint:   tripdomain.Point{Latitude: 59.929, Longitude: 30.3626},
		Price:      1450,
	}

	trip, err := usecases.create.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != tripdomain.StatusActive || trip.FinishedAt != nil {
		t.Fatalf("created trip has incorrect status: %+v", trip)
	}
	got, err := usecases.get.Get(ctx, trip.ID)
	if err != nil || got.ID != trip.ID || got.Price != input.Price {
		t.Fatalf("get trip: got %+v, error %v", got, err)
	}
	if _, err := usecases.create.Create(ctx, input); !errors.Is(err, tripdomain.ErrDriverBusy) {
		t.Fatalf("duplicate active trip: got %v, want tripdomain.ErrDriverBusy", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, err := usecases.finish.Finish(ctx, trip.ID)
			results <- err
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	var succeeded, completed int
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, tripdomain.ErrCompleted):
			completed++
		default:
			t.Fatalf("unexpected finish result: %v", err)
		}
	}
	if succeeded != 1 || completed != 1 {
		t.Fatalf("concurrent finish: got %d success, %d completed", succeeded, completed)
	}
	got, err = usecases.get.Get(ctx, trip.ID)
	if err != nil || got.Status != tripdomain.StatusCompleted || got.FinishedAt == nil {
		t.Fatalf("finished trip: got %+v, error %v", got, err)
	}
	var historyCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", trip.ID).Scan(&historyCount); err != nil || historyCount != 2 {
		t.Fatalf("history rows: got %d, error %v", historyCount, err)
	}
	if _, err := usecases.create.Create(ctx, input); err != nil {
		t.Fatalf("driver should be free after finish: %v", err)
	}

	busyDriverID := uuid.New()
	cleanupDriver(t, pool, busyDriverID)
	input.DriverID = busyDriverID
	start = make(chan struct{})
	results = make(chan error, 20)
	for range 20 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, err := usecases.create.Create(ctx, input)
			results <- err
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	succeeded, busy := 0, 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, tripdomain.ErrDriverBusy):
			busy++
		default:
			t.Fatalf("unexpected create result: %v", err)
		}
	}
	if succeeded != 1 || busy != 19 {
		t.Fatalf("concurrent create: got %d success, %d driver busy", succeeded, busy)
	}
}

func TestCreateRollsBackWhenHistoryFails(t *testing.T) {
	usecases, pool := testUsecases(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	driverID := uuid.New()
	cleanupDriver(t, pool, driverID)

	name := "trip_history_fail_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	functionName := name + "_fn"
	functionSQL := fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM trips WHERE id = NEW.trip_id AND driver_id = '%s') THEN
        RAISE EXCEPTION 'forced history failure';
    END IF;
    RETURN NEW;
END $$`, functionName, driverID)
	if _, err := pool.Exec(ctx, functionSQL); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, fmt.Sprintf("CREATE TRIGGER %s BEFORE INSERT ON trip_status_history FOR EACH ROW EXECUTE FUNCTION %s()", name, functionName)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DROP TRIGGER "+name+" ON trip_status_history"); err != nil {
			t.Errorf("drop trigger: %v", err)
		}
		if _, err := pool.Exec(cleanupCtx, "DROP FUNCTION "+functionName+"()"); err != nil {
			t.Errorf("drop trigger function: %v", err)
		}
	})

	_, err := usecases.create.Create(ctx, model.CreateTripInput{
		UserID: uuid.New(), DriverID: driverID,
		StartPoint: tripdomain.Point{Latitude: 1, Longitude: 2},
		EndPoint:   tripdomain.Point{Latitude: 3, Longitude: 4},
		Price:      100,
	})
	if err == nil {
		t.Fatal("create succeeded despite failed history insert")
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trips WHERE driver_id = $1", driverID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("trip survived rollback: count %d, error %v", count, err)
	}
}

func TestCreateWithKey(t *testing.T) {
	usecases, pool := testUsecases(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	driverID := uuid.New()
	cleanupDriver(t, pool, driverID)
	input := model.CreateTripInput{
		UserID: uuid.New(), DriverID: driverID,
		StartPoint: tripdomain.Point{Latitude: 1, Longitude: 2},
		EndPoint:   tripdomain.Point{Latitude: 3, Longitude: 4}, Price: 100,
	}
	key := uuid.New()
	start := make(chan struct{})
	type result struct {
		trip     tripdomain.Trip
		replayed bool
		err      error
	}
	results := make(chan result, 12)
	var workers sync.WaitGroup
	for range 12 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			trip, replayed, err := usecases.create.CreateWithKey(ctx, input, &key)
			results <- result{trip, replayed, err}
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	var firstID uuid.UUID
	var created, replayed int
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent keyed create: %v", result.err)
		}
		if firstID == uuid.Nil {
			firstID = result.trip.ID
		} else if result.trip.ID != firstID {
			t.Fatalf("key produced different trips: %s and %s", firstID, result.trip.ID)
		}
		if result.replayed {
			replayed++
		} else {
			created++
		}
	}
	if created != 1 || replayed != 11 {
		t.Fatalf("got %d creates and %d replays", created, replayed)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trips WHERE driver_id = $1", driverID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("saved trips: %d, %v", count, err)
	}

	changed := input
	changed.Price++
	if _, _, err := usecases.create.CreateWithKey(ctx, changed, &key); !errors.Is(err, tripdomain.ErrIdempotencyConflict) {
		t.Fatalf("changed request: got %v, want idempotency conflict", err)
	}
	if _, err := usecases.finish.Finish(ctx, firstID); err != nil {
		t.Fatal(err)
	}
	snapshot, wasReplayed, err := usecases.create.CreateWithKey(ctx, input, &key)
	if err != nil || !wasReplayed || snapshot.ID != firstID || snapshot.Status != tripdomain.StatusActive {
		t.Fatalf("replay after finish: %+v, replayed %t, error %v", snapshot, wasReplayed, err)
	}

	if _, err := pool.Exec(ctx, "UPDATE trip_idempotency_keys SET expires_at = now() - interval '1 second' WHERE key = $1", key); err != nil {
		t.Fatal(err)
	}
	newTrip, wasReplayed, err := usecases.create.CreateWithKey(ctx, changed, &key)
	if err != nil || wasReplayed || newTrip.ID == firstID {
		t.Fatalf("expired key: %+v, replayed %t, error %v", newTrip, wasReplayed, err)
	}
}

type tripUsecases struct {
	create *create.Usecase
	get    *get.Usecase
	finish *finish.Usecase
}

func testUsecases(t *testing.T) (tripUsecases, *pgxpool.Pool) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	txManager := transaction.NewTxManager(pool, 5*time.Second)
	repo := triprepository.NewRepository(txManager, 3*time.Second)
	return tripUsecases{
		create: create.New(repo, txManager),
		get:    get.New(repo),
		finish: finish.New(repo, txManager),
	}, pool
}

func cleanupDriver(t *testing.T, pool *pgxpool.Pool, driverID uuid.UUID) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := pool.Exec(ctx, "DELETE FROM trips WHERE driver_id = $1", driverID); err != nil {
			t.Errorf("delete test trips: %v", err)
		}
	})
}
