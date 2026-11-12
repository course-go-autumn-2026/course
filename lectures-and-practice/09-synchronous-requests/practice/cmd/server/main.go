package main

import (
	pb "example.com/sync-practice/api/delivery/v1"
	"example.com/sync-practice/task01"
	p1 "example.com/sync-practice/task01/api"
	"example.com/sync-practice/task02"
	"example.com/sync-practice/task03"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

func main() {
	task := flag.Int("task", 1, "1, 2 or 3")
	addr := flag.String("addr", "127.0.0.1:50051", "listen address")
	delay := flag.Duration("delay", 2*time.Second, "task 3 calculation delay")
	flag.Parse()
	l, e := net.Listen("tcp", *addr)
	if e != nil {
		log.Fatal(e)
	}
	s := grpc.NewServer()
	switch *task {
	case 1:
		p1.RegisterDeliveryServiceServer(s, &task01.Server{})
	case 2:
		pb.RegisterDeliveryServiceServer(s, &task02.Server{})
	case 3:
		pb.RegisterDeliveryServiceServer(s, &task03.Server{Delay: *delay, OnStart: func() { log.Print("calculation started") }, OnDone: func(e error) { log.Printf("calculation ended: %v", e) }})
	default:
		log.Fatal("task must be 1, 2 or 3")
	}
	log.Printf("task=%d listening on %s", *task, l.Addr())
	if e = s.Serve(l); e != nil {
		log.Fatal(e)
	}
}
