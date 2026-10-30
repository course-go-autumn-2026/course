package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"example.com/food-observability/internal/httpserver"
	"example.com/food-observability/telemetry"
)

func main() {
	if err := run(os.Args[1:]); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) (err error) {
	flags := flag.NewFlagSet("couriers", flag.ContinueOnError)
	port := flags.Int("port", 8081, "HTTP port on 127.0.0.1")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *port < 1 || *port > 65535 {
		return errors.New("expected flags only and a port between 1 and 65535")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger, shutdown, err := telemetry.Setup(ctx, "courier-service")
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = errors.Join(err, shutdown(flushCtx))
	}()

	return httpserver.Serve(ctx, "127.0.0.1:"+strconv.Itoa(*port), couriersHandler(logger), logger)
}
