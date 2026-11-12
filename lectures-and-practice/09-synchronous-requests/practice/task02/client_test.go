package task02

import (
	"context"
	"errors"
	pb "example.com/sync-practice/api/delivery/v1"
	"example.com/sync-practice/internal/domain"
	"example.com/sync-practice/internal/testserver"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestMapError(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code codes.Code
	}{
		{nil, codes.OK}, {fmt.Errorf("wrapped: %w", domain.ErrInvalid), codes.InvalidArgument},
		{domain.ErrOutside, codes.NotFound}, {domain.ErrUnavailable, codes.Unavailable},
		{errors.New("secret database password"), codes.Internal},
	} {
		got := MapError(tc.err)
		if status.Code(got) != tc.code {
			t.Errorf("%v: got %v want %v", tc.err, got, tc.code)
		}
	}
	if status.Convert(MapError(errors.New("secret"))).Message() == "secret" {
		t.Fatal("утечка внутренних деталей")
	}
}
func TestClient(t *testing.T) {
	addr := testserver.Start(t, func(s *grpc.Server) { pb.RegisterDeliveryServiceServer(s, &Server{}) })
	c, err := NewClient(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, tc := range []struct {
		addr string
		w    int64
		want error
	}{
		{"Центр", 1500, nil}, {"", 1500, ErrInput}, {"Центр", 0, ErrInput}, {"Центр", 20001, ErrInput},
		{"Луна", 1500, ErrNoDelivery}, {"Недоступно", 1500, ErrDependency},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		out, err := c.Estimate(ctx, tc.addr, tc.w)
		cancel()
		if !errors.Is(err, tc.want) {
			t.Errorf("%q: got %v want %v", tc.addr, err, tc.want)
		}
		if tc.want == nil && (out == nil || out.GetPriceKopecks() != 30000 || out.GetEtaMinutes() != 30) {
			t.Errorf("неверный ответ: %v", out)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.Estimate(ctx, "Центр", 1500)
	if status.Code(err) != codes.Canceled {
		t.Fatalf("потерян контекст: %v", err)
	}
}
