package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
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
	flags := flag.NewFlagSet("orders", flag.ContinueOnError)
	port := flags.Int("port", 8080, "HTTP port on 127.0.0.1")
	courierURL := flags.String("courier", "http://127.0.0.1:8081/reserve", "courier reservation URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *port < 1 || *port > 65535 {
		return errors.New("expected flags only and a port between 1 and 65535")
	}
	u, err := url.Parse(*courierURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("courier URL must be HTTP(S), without credentials, query or fragment")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger, shutdown, err := telemetry.Setup(ctx, "order-service")
	if err != nil {
		return err
	}
	defer func() {
		// Signal cancellation must not cancel the final telemetry export.
		flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = errors.Join(err, shutdown(flushCtx))
	}()

	client := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()
	handler, err := ordersHandler(*courierURL, client, logger)
	if err != nil {
		return err
	}
	return httpserver.Serve(ctx, "127.0.0.1:"+strconv.Itoa(*port), handler, logger)
}
