package generator

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"PythonToGoMigrationTest/internal/config"
)

func Start() {
	if err := run(); err != nil {
		slog.Error("generator failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.NewGeneratorConfig()
	if err != nil {
		return err
	}

	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}
	if (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" {
		return errors.New("generator URL must be an absolute HTTP or HTTPS URL")
	}
	timeout := time.Duration(cfg.Timeout * float64(time.Second))
	if timeout <= 0 {
		return errors.New("timeout must be positive and fit within time.Duration")
	}
	interval := time.Duration(cfg.Interval * float64(time.Second))
	client := &http.Client{Timeout: timeout}
	defer client.CloseIdleConnections()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stats := &Stats{}
	var workers sync.WaitGroup
	for i := 0; i < cfg.Threads; i++ {
		workers.Add(1)
		go func(workerID int) {
			defer workers.Done()
			workerJob(ctx, workerID, client, baseURL, interval, stats)
		}(i)
	}
	slog.Info("generator started", "threads", cfg.Threads, "url", cfg.URL)

	<-ctx.Done()
	stop()
	slog.Info("stopping generator")
	workers.Wait()
	stats.PrintTotals()
	return nil
}
