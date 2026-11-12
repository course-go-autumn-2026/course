package task01

import (
	"context"
	"example.com/sync-practice/internal/testserver"
	pb "example.com/sync-practice/task01/api"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestContract(t *testing.T) {
	d := (&pb.EstimateResponse{}).ProtoReflect().Descriptor()
	f := d.Fields().ByName("eta_minutes")
	if f == nil {
		t.Fatal("добавьте eta_minutes в proto и выполните генерацию")
	}
	if f.Number() != 2 || f.Kind().String() != "int32" {
		t.Fatal("eta_minutes: ожидается int32 и номер 2")
	}
}
func TestEstimate(t *testing.T) {
	addr := testserver.Start(t, func(s *grpc.Server) { pb.RegisterDeliveryServiceServer(s, &Server{}) })
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewDeliveryServiceClient(conn)
	for _, tc := range []struct {
		address       string
		weight, price int64
		eta           int64
	}{
		{"Центр", 1500, 30000, 30}, {"Север", 1500, 40000, 45}, {"Центр", 1000, 25000, 30}, {"Центр", 1001, 30000, 30},
	} {
		t.Run(fmt.Sprintf("%s/%dg", tc.address, tc.weight), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			out, err := client.Estimate(ctx, &pb.EstimateRequest{Address: tc.address, WeightGrams: tc.weight})
			if err != nil {
				t.Fatal(err)
			}
			if out.GetPriceKopecks() != tc.price {
				t.Fatalf("price=%d want %d", out.GetPriceKopecks(), tc.price)
			}
			f := out.ProtoReflect().Descriptor().Fields().ByName("eta_minutes")
			if f == nil {
				t.Fatal("нет eta_minutes")
			}
			if got := out.ProtoReflect().Get(f).Int(); got != tc.eta {
				t.Fatalf("eta=%d want %d", got, tc.eta)
			}
		})
	}
}
