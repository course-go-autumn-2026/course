package task03

import (
	"context"
	pb "example.com/sync-practice/api/delivery/v1"
	"time"
)

func Estimate(ctx context.Context, client pb.DeliveryServiceClient, in *pb.EstimateRequest, budget time.Duration) (*pb.EstimateResponse, error) {
	// TODO: создайте дочерний контекст с budget, вызовите cancel и передайте контекст в RPC.
	return client.Estimate(ctx, in)
}
