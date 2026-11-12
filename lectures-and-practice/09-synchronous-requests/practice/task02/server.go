package task02

import (
	"context"
	pb "example.com/sync-practice/api/delivery/v1"
	"example.com/sync-practice/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedDeliveryServiceServer
}

func (s *Server) Estimate(ctx context.Context, in *pb.EstimateRequest) (*pb.EstimateResponse, error) {
	q, err := domain.Estimate(in.GetAddress(), in.GetWeightGrams())
	if err != nil {
		return nil, MapError(err)
	}
	return &pb.EstimateResponse{PriceKopecks: q.PriceKopecks, EtaMinutes: q.ETAMinutes}, nil
}
func MapError(err error) error {
	if err == nil {
		return nil
	}
	// TODO: errors.Is, согласованные статусы. Не передавайте внутренний текст клиенту.
	return status.Error(codes.Internal, "estimate failed")
}
