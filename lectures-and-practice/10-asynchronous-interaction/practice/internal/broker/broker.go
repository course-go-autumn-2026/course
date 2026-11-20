package broker

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"net"
	"os"
	"strconv"
	"time"
)

func Address() string {
	if a := os.Getenv("KAFKA_ADDR"); a != "" {
		return a
	}
	return "localhost:19092"
}
func Writer(topic string) *kafka.Writer {
	return &kafka.Writer{Addr: kafka.TCP(Address()), Topic: topic, Balancer: &kafka.Hash{}, RequiredAcks: kafka.RequireAll, Async: false, WriteTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, MaxAttempts: 3, BatchTimeout: 10 * time.Millisecond}
}
func Reader(topic, group string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{Brokers: []string{Address()}, Topic: topic, GroupID: group, StartOffset: kafka.FirstOffset, CommitInterval: 0, MaxWait: 200 * time.Millisecond, MinBytes: 1, MaxBytes: 1e6})
}
func Setup(ctx context.Context) error {
	c, err := kafka.DialContext(ctx, "tcp", Address())
	if err != nil {
		return err
	}
	defer c.Close()
	controller, err := c.Controller()
	if err != nil {
		return err
	}
	admin, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer admin.Close()
	var configs []kafka.TopicConfig
	for _, topic := range []string{"orders.task1", "orders.task2", "orders.task3", "orders.dlq"} {
		configs = append(configs, kafka.TopicConfig{Topic: topic, NumPartitions: 3, ReplicationFactor: 1})
	}
	if err = admin.CreateTopics(configs...); err != nil {
		return fmt.Errorf("create topics: %w", err)
	}
	return nil
}
