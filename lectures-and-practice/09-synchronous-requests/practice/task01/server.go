package task01

import (
	"context"
	pb "example.com/sync-practice/task01/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedDeliveryServiceServer
}

func (s *Server) Estimate(ctx context.Context, in *pb.EstimateRequest) (*pb.EstimateResponse, error) {
	// TODO: вызовите domain.Estimate и преобразуйте Quote в protobuf-ответ.
	// В задании 1 проверяются только успешные запросы. Ошибки разбираем в задании 2.
	return nil, status.Error(codes.Unimplemented, "task 1")
}
