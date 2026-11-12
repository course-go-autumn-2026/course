package task02

import (
	"context"
	"errors"
	pb "example.com/sync-practice/api/delivery/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrInput      = errors.New("проверьте адрес и вес")
	ErrNoDelivery = errors.New("адрес вне зоны доставки")
	ErrDependency = errors.New("служба доставки недоступна")
)

type Client struct {
	conn *grpc.ClientConn
	rpc  pb.DeliveryServiceClient
}

func NewClient(target string) (*Client, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, rpc: pb.NewDeliveryServiceClient(conn)}, nil
}
func (c *Client) Close() error { return c.conn.Close() }
func (c *Client) Estimate(ctx context.Context, address string, weight int64) (*pb.EstimateResponse, error) {
	// TODO: вызовите c.rpc.Estimate с ctx и преобразуйте три согласованных статуса.
	// Остальные ошибки возвращайте как есть, сохраняя gRPC-код.
	return nil, errors.New("TODO: task 2 client")
}
