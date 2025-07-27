package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Adit0507/url-shortener/bloom"
	"github.com/Adit0507/url-shortener/handler"
	"github.com/Adit0507/url-shortener/snowflake"
	"github.com/Adit0507/url-shortener/store"
)

func main() {
    // Initialize components
    snowflakeGen := snowflake.NewSnowFlake(1)
    bloomFilter := bloom.NewBloomFilter(1000000, 0.01)
    urlStore := store.NewURLStore()

    // Initialize worker pool for storing URLs
    taskChan := make(chan store.Task, 100)
    var wg sync.WaitGroup
    const numWorkers = 4
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range taskChan {
                urlStore.Store(task.ShortCode, task.URL)
                bloomFilter.Add(task.ShortCode)
            }
        }()
    }

    // Set up HTTP server
    server := &http.Server{Addr: ":8080"}
    http.HandleFunc("/shorten", handler.ShortenHandler(snowflakeGen, bloomFilter, urlStore, taskChan))
    http.HandleFunc("/", handler.RedirectHandler(urlStore))

    // Start server in a goroutine
    go func() {
        fmt.Println("Starting server on :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // Shutdown server
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Shutdown error: %v", err)
    }

    // Close task channel and wait for workers
    close(taskChan)
    wg.Wait()
    fmt.Println("Server shut down gracefully")
}