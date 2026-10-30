package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	endpoint := flag.String("url", "http://127.0.0.1:8080/orders", "URL for empty POST requests")
	rps := flag.Int("rps", 5, "target requests per second")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, *endpoint, *rps, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

type result struct {
	id       uint64
	status   int
	duration time.Duration
	err      error
}

func run(ctx context.Context, endpoint string, rps int, output io.Writer) error {
	if rps < 1 || rps > int(time.Second) {
		return fmt.Errorf("rps must be between 1 and %d", int(time.Second))
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	if (request.URL.Scheme != "http" && request.URL.Scheme != "https") || request.URL.Host == "" {
		return fmt.Errorf("URL must use http or https and include a host")
	}
	client := &http.Client{
		Timeout: 4 * time.Second,
		// Like curl without -L: a redirect is one response, not another request.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	defer client.CloseIdleConnections()
	logger := log.New(output, "", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("event=started method=POST url=%q body=empty target_rps=%d timeout=4s stop=Ctrl+C", endpoint, rps)
	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()
	stats := time.NewTicker(time.Second)
	defer stats.Stop()
	results := make(chan result)
	var requests, inFlight, success, failed, cancelled uint64
	lastTime, lastRequests := time.Now(), uint64(0)
	report := func(event string) {
		now := time.Now()
		logger.Printf("event=%s requests=%d completed=%d success=%d error=%d cancelled=%d in_flight=%d actual_rps=%.1f target_rps=%d",
			event, requests, requests-inFlight, success, failed, cancelled, inFlight,
			float64(requests-lastRequests)/now.Sub(lastTime).Seconds(), rps)
		lastTime, lastRequests = now, requests
	}

	ticks, done := ticker.C, ctx.Done()
	for ticks != nil || inFlight > 0 {
		select {
		case <-done:
			ticker.Stop()
			ticks, done = nil, nil // Stop sending, then collect cancelled in-flight requests.
		case <-ticks:
			if ctx.Err() != nil {
				continue
			}
			requests++
			inFlight++
			logger.Printf("request=%d event=send method=POST url=%q body=empty", requests, endpoint)
			// ponytail: one goroutine per tick; the 4s timeout limits overlap at 5 RPS.
			// Add a worker cap if much higher rates are needed.
			go func(id uint64) {
				time.Sleep(time.Duration(10+rand.IntN(191)) * time.Millisecond)
				started := time.Now()
				response, err := client.Do(request.Clone(ctx))
				status := 0
				if err == nil {
					status = response.StatusCode
					_, err = io.Copy(io.Discard, response.Body)
					response.Body.Close()
				}
				results <- result{id: id, status: status, duration: time.Since(started), err: err}
			}(requests)
		case result := <-results:
			inFlight--
			outcome := "error"
			switch {
			case result.err != nil && ctx.Err() != nil && errors.Is(result.err, ctx.Err()):
				cancelled++
				outcome = "cancelled"
			case result.err == nil && result.status >= 200 && result.status < 300:
				success++
				outcome = "success"
			default:
				failed++
			}
			detail := ""
			if result.err != nil {
				detail = fmt.Sprintf(" error=%q", result.err)
			}
			logger.Printf("request=%d event=done status=%d duration=%s outcome=%s%s",
				result.id, result.status, result.duration.Round(time.Microsecond), outcome, detail)
		case <-stats.C:
			report("progress")
		}
	}
	report("stopped")
	return nil
}
