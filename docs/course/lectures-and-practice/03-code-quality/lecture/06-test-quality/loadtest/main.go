// Команда loadtest отправляет конкурентные HTTP-запросы и выводит основные метрики нагрузки.
package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type result struct {
	duration time.Duration
	failed   bool
}

func main() {
	target := flag.String("url", "http://localhost:8080/greet?name=Load", "target URL")
	requests := flag.Int("requests", 1000, "total number of requests")
	concurrency := flag.Int("concurrency", 20, "number of concurrent workers")
	flag.Parse()

	if *requests <= 0 || *concurrency <= 0 {
		fmt.Fprintln(os.Stderr, "requests and concurrency must be positive")
		os.Exit(2)
	}

	results := run(*target, *requests, *concurrency)
	printReport(results)
}

func run(target string, requestCount, concurrency int) []result {
	client := &http.Client{Timeout: 3 * time.Second}
	jobs := make(chan struct{}, requestCount)
	results := make(chan result, requestCount)

	for range requestCount {
		jobs <- struct{}{}
	}
	close(jobs)

	var workers sync.WaitGroup
	for range concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range jobs {
				started := time.Now()
				response, err := client.Get(target)
				item := result{duration: time.Since(started), failed: err != nil}
				if err == nil {
					_, copyErr := io.Copy(io.Discard, response.Body)
					closeErr := response.Body.Close()
					item.failed = response.StatusCode >= http.StatusBadRequest || copyErr != nil || closeErr != nil
				}
				results <- item
			}
		}()
	}

	workers.Wait()
	close(results)

	collected := make([]result, 0, requestCount)
	for item := range results {
		collected = append(collected, item)
	}
	return collected
}

func printReport(results []result) {
	durations := make([]time.Duration, 0, len(results))
	var failed int
	var total time.Duration
	for _, item := range results {
		durations = append(durations, item.duration)
		total += item.duration
		if item.failed {
			failed++
		}
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	p95Index := int(math.Ceil(float64(len(durations))*0.95)) - 1
	average := total / time.Duration(len(durations))

	fmt.Printf("requests=%d errors=%d average=%s p95=%s\n",
		len(results), failed, average.Round(time.Microsecond), durations[p95Index].Round(time.Microsecond))
}
