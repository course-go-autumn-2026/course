package main

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/async-practice/internal/broker"
	"example.com/async-practice/internal/event"
	"example.com/async-practice/internal/processing"
	"example.com/async-practice/internal/replay"
	"example.com/async-practice/task01"
	"example.com/async-practice/task02"
	"example.com/async-practice/task03"
	"flag"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	mode := flag.String("mode", "send", "setup/send/consume/dlq/replay")
	task := flag.Int("task", 1, "1..3")
	id := flag.String("id", "event-1", "stable event ID")
	order := flag.String("order", "order-1", "order ID and key")
	address := flag.String("address", "Центр", "address")
	group := flag.String("group", "delivery-demo", "consumer group")
	scenario := flag.String("scenario", "ok", "ok/transient/permanent/always-fail")
	count := flag.Int("count", 1, "messages to read; 0 until Ctrl+C")
	crash := flag.Bool("crash-before-commit", false, "stop after handle/DLQ, before offset commit")
	dlqFail := flag.Bool("dlq-fail", false, "simulate DLQ write error")
	partition := flag.Int("partition", -1, "exact DLQ partition to replay")
	offset := flag.Int64("offset", -1, "exact DLQ offset to replay")
	flag.Parse()
	if *task < 1 || *task > 3 || *count < 0 {
		return errors.New("invalid task/count")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	topic := fmt.Sprintf("orders.task%d", *task)
	switch *mode {
	case "setup":
		c, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := broker.Setup(c); err != nil {
			return err
		}
		log.Print("topics ready")
		return nil
	case "send":
		e := event.OrderCreated{ID: *id, OrderID: *order, Address: *address, Version: 1}
		w := broker.Writer(topic)
		defer w.Close()
		if *task == 1 {
			if err := task01.Publish(ctx, w, e); err != nil {
				return err
			}
		} else {
			if err := e.Validate(); err != nil {
				return err
			}
			b, err := json.Marshal(e)
			if err != nil {
				return err
			}
			if err = w.WriteMessages(ctx, kafka.Message{Key: []byte(e.OrderID), Value: b}); err != nil {
				return err
			}
		}
		log.Printf("published id=%s key=%s topic=%s", e.ID, e.OrderID, topic)
		return nil
	case "replay":
		if *partition < 0 || *offset < 0 {
			return errors.New("replay requires explicit -partition and -offset from DLQ listing")
		}
		c, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		rd := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{broker.Address()}, Topic: "orders.dlq", Partition: *partition, MinBytes: 1, MaxBytes: 1e6})
		defer rd.Close()
		if err := rd.SetOffset(*offset); err != nil {
			return err
		}
		m, err := rd.FetchMessage(c)
		if err != nil {
			return err
		}
		if m.Offset != *offset {
			return errors.New("requested DLQ offset no longer available")
		}
		w := broker.Writer("orders.task3")
		defer w.Close()
		if err = replay.Publish(c, w, m.Value); err != nil {
			return err
		}
		log.Printf("replayed DLQ partition=%d offset=%d; original ID/key/value preserved", m.Partition, m.Offset)
		return nil
	case "dlq":
		rd := broker.Reader("orders.dlq", *group)
		defer rd.Close()
		for n := 0; *count == 0 || n < *count; n++ {
			m, err := rd.FetchMessage(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("DLQ partition=%d offset=%d\n%s\n", m.Partition, m.Offset, m.Value)
			if err = rd.CommitMessages(ctx, m); err != nil {
				return err
			}
		}
		return nil
	case "consume":
		if *scenario != "ok" && *scenario != "transient" && *scenario != "permanent" && *scenario != "always-fail" {
			return errors.New("unknown scenario")
		}
		rd := broker.Reader(topic, *group)
		defer rd.Close()
		dw := broker.Writer("orders.dlq")
		defer dw.Close()
		attempts := map[string]int{}
		handle := func(c context.Context, m kafka.Message) error {
			log.Printf("received topic=%s partition=%d offset=%d key=%s", m.Topic, m.Partition, m.Offset, m.Key)
			var e event.OrderCreated
			if err := json.Unmarshal(m.Value, &e); err != nil {
				return fmt.Errorf("%w: invalid JSON", processing.ErrPermanent)
			}
			if err := e.Validate(); err != nil {
				return fmt.Errorf("%w: %v", processing.ErrPermanent, err)
			}
			attempts[e.ID]++
			log.Printf("attempt id=%s n=%d", e.ID, attempts[e.ID])
			if *scenario == "permanent" {
				return fmt.Errorf("%w: simulated unsupported address", processing.ErrPermanent)
			}
			if *scenario == "always-fail" || (*scenario == "transient" && attempts[e.ID] < 3) {
				return errors.New("temporary simulator failure")
			}
			log.Printf("handled id=%s attempt=%d", e.ID, attempts[e.ID])
			return nil
		}
		commit := func(c context.Context, m kafka.Message) error {
			if *crash {
				return errors.New("simulated crash before offset commit")
			}
			if err := rd.CommitMessages(c, m); err != nil {
				return err
			}
			log.Printf("offset committed partition=%d offset=%d", m.Partition, m.Offset)
			return nil
		}
		for n := 0; *count == 0 || n < *count; n++ {
			var err error
			if *task == 1 {
				err = task01.ConsumeOne(ctx, rd, handle)
			} else {
				var m kafka.Message
				m, err = rd.FetchMessage(ctx)
				if err != nil {
					return err
				}
				if *task == 2 {
					err = task02.Process(ctx, m, task02.Hooks{Handle: handle, Wait: processing.Wait, Commit: commit})
				} else {
					err = task03.Process(ctx, m, task03.Hooks{Handle: handle, Commit: commit, DeadLetter: func(c context.Context, d task03.DLQ) error {
						if *dlqFail {
							return errors.New("simulated DLQ outage")
						}
						b, err := json.Marshal(d)
						if err != nil {
							return err
						}
						err = dw.WriteMessages(c, kafka.Message{Key: d.Key, Value: b})
						if err == nil {
							log.Print("DLQ acknowledged")
						}
						return err
					}})
				}
			}
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("unknown mode")
	}
}
