package task03

import (
	"context"
	pb "example.com/sync-practice/api/delivery/v1"
	"example.com/sync-practice/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Server struct {
	pb.UnimplementedDeliveryServiceServer
	Delay time.Duration
	// OnStart/OnDone нужны стенду и тестам, вызываются синхронно внутри handler.
	OnStart func()
	OnDone  func(error)
}

func (s *Server) Estimate(ctx context.Context, in *pb.EstimateRequest) (*pb.EstimateResponse, error) {
	if s.OnStart != nil {
		s.OnStart()
	}
	err := Wait(ctx, s.Delay)
	if s.OnDone != nil {
		s.OnDone(err)
	}
	if err != nil {
		return nil, status.FromContextError(err).Err()
	}
	q, err := domain.Estimate(in.GetAddress(), in.GetWeightGrams())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "use valid demo input")
	}
	return &pb.EstimateResponse{PriceKopecks: q.PriceKopecks, EtaMinutes: q.ETAMinutes}, nil
}
