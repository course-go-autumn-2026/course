package task03

import (
	"context"
	"errors"
	pb "example.com/sync-practice/api/delivery/v1"
	"example.com/sync-practice/internal/testserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func fixture(t *testing.T, delay time.Duration) (pb.DeliveryServiceClient, <-chan struct{}, <-chan error) {
	started := make(chan struct{}, 1)
	done := make(chan error, 1)
	addr := testserver.Start(t, func(s *grpc.Server) {
		pb.RegisterDeliveryServiceServer(s, &Server{Delay: delay, OnStart: func() { started <- struct{}{} }, OnDone: func(e error) { done <- e }})
	})
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return pb.NewDeliveryServiceClient(conn), started, done
}

var request = &pb.EstimateRequest{Address: "Центр", WeightGrams: 1500}

func TestDeadline(t *testing.T) {
	c, started, done := fixture(t, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	_, err := Estimate(ctx, c, request, 200*time.Millisecond)
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("got %v, want DeadlineExceeded", err)
	}
	select {
	case <-started:
	default:
		t.Fatal("вызов не дошёл до сервера; повторите на свободной машине")
	}
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) && !errors.Is(e, context.DeadlineExceeded) {
			t.Fatalf("server work: %v", e)
		}
	case <-time.After(700 * time.Millisecond):
		t.Fatal("клиент завершился, но сервер продолжает работу")
	}
}
func TestExplicitCancellation(t *testing.T) {
	c, started, done := fixture(t, 2*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, e := Estimate(ctx, c, request, 4*time.Second); result <- e }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("сервер не начал работу")
	}
	cancel()
	select {
	case e := <-result:
		if status.Code(e) != codes.Canceled {
			t.Fatalf("got %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("клиент игнорирует отмену")
	}
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatalf("сервер: %v", e)
		}
	case <-time.After(700 * time.Millisecond):
		t.Fatal("сервер игнорирует отмену")
	}
}
func TestFastRequest(t *testing.T) {
	c, _, _ := fixture(t, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := Estimate(ctx, c, request, time.Second)
	if err != nil || out.GetPriceKopecks() != 30000 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
func TestParentBudget(t *testing.T) {
	c, _, _ := fixture(t, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := Estimate(ctx, c, request, time.Second)
	if status.Code(err) != codes.DeadlineExceeded || time.Since(start) > 700*time.Millisecond {
		t.Fatalf("потерян deadline родителя: %v", err)
	}
}
