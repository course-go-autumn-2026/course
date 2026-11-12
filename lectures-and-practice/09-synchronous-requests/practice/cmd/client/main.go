package main

import (
	"context"
	pb "example.com/sync-practice/api/delivery/v1"
	p1 "example.com/sync-practice/task01/api"
	"example.com/sync-practice/task02"
	"example.com/sync-practice/task03"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		if st, ok := status.FromError(err); ok {
			fmt.Fprintf(os.Stderr, "error: %s (grpc=%s)\n", st.Message(), st.Code())
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}
func run() error {
	task := flag.Int("task", 1, "1, 2 or 3")
	addr := flag.String("addr", "127.0.0.1:50051", "server")
	address := flag.String("address", "Центр", "delivery address")
	weight := flag.Int64("weight", 1500, "grams")
	budget := flag.Duration("timeout", 200*time.Millisecond, "task 3 RPC budget")
	cancelAfter := flag.Duration("cancel-after", 0, "explicit cancellation after duration")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if *cancelAfter > 0 {
		timer := time.AfterFunc(*cancelAfter, cancel)
		defer timer.Stop()
	}
	if *task == 2 {
		c, e := task02.NewClient(*addr)
		if e != nil {
			return e
		}
		defer c.Close()
		out, e := c.Estimate(ctx, *address, *weight)
		if e != nil {
			return e
		}
		fmt.Println(out)
		return nil
	}
	conn, e := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if e != nil {
		return e
	}
	defer conn.Close()
	switch *task {
	case 1:
		out, e := p1.NewDeliveryServiceClient(conn).Estimate(ctx, &p1.EstimateRequest{Address: *address, WeightGrams: *weight})
		if e != nil {
			return e
		}
		fmt.Println(out)
	case 3:
		out, e := task03.Estimate(ctx, pb.NewDeliveryServiceClient(conn), &pb.EstimateRequest{Address: *address, WeightGrams: *weight}, *budget)
		if e != nil {
			return e
		}
		fmt.Println(out)
	default:
		return fmt.Errorf("task must be 1, 2 or 3")
	}
	return nil
}
