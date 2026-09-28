package trips

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestServiceConcurrencyAndHistory(t *testing.T) {
	service, pool := testService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID, driverID := uuid.New(), uuid.New()
	cleanupDriver(t, pool, driverID)
	input := CreateInput{
		UserID: userID, DriverID: driverID,
		StartPoint: Point{Latitude: 59.9398, Longitude: 30.3146},
		EndPoint:   Point{Latitude: 59.929, Longitude: 30.3626},
		Price:      1450,
	}

	trip, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != StatusActive || trip.FinishedAt != nil {
		t.Fatalf("created trip has incorrect status: %+v", trip)
	}
	got, err := service.Get(ctx, trip.ID)
	if err != nil || got.ID != trip.ID || got.Price != input.Price {
		t.Fatalf("get trip: got %+v, error %v", got, err)
	}
	if _, err := service.Create(ctx, input); !errors.Is(err, ErrDriverBusy) {
		t.Fatalf("duplicate active trip: got %v, want ErrDriverBusy", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, err := service.Finish(ctx, trip.ID)
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
		case errors.Is(err, ErrCompleted):
			completed++
		default:
			t.Fatalf("unexpected finish result: %v", err)
		}
	}
	if succeeded != 1 || completed != 1 {
		t.Fatalf("concurrent finish: got %d success, %d completed", succeeded, completed)
	}
	got, err = service.Get(ctx, trip.ID)
	if err != nil || got.Status != StatusCompleted || got.FinishedAt == nil {
		t.Fatalf("finished trip: got %+v, error %v", got, err)
	}
	var historyCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM trip_status_history WHERE trip_id = $1", trip.ID).Scan(&historyCount); err != nil || historyCount != 2 {
		t.Fatalf("history rows: got %d, error %v", historyCount, err)
	}
	if _, err := service.Create(ctx, input); err != nil {
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
			_, err := service.Create(ctx, input)
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
		case errors.Is(err, ErrDriverBusy):
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
	service, pool := testService(t)
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

	_, err := service.Create(ctx, CreateInput{
		UserID: uuid.New(), DriverID: driverID,
		StartPoint: Point{Latitude: 1, Longitude: 2},
		EndPoint:   Point{Latitude: 3, Longitude: 4},
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

func testService(t *testing.T) (*Service, *pgxpool.Pool) {
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
	txManager := database.NewTxManager(pool, 5*time.Second)
	repo := NewRepository(txManager, 3*time.Second)
	return NewService(repo, txManager), pool
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
