package calculator

import (
	"PythonToGoMigrationTest/internal/config"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func Start() {
	cfg, err := config.NewCalculatorConfig()

	if err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}

	interval := time.Duration(cfg.Interval * float64(time.Second))
	if interval <= 0 {
		slog.Error("interval must be positive and fit within time.Duration", "interval", cfg.Interval)
		os.Exit(1)
	}

	loader := NewLibLoader(cfg)
	cLib, rustLib, err := loader.load()

	if err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}

	state := NewState(cLib, rustLib)
	mux := http.NewServeMux()
	mux.Handle("POST /calc", NewCalcHandler(state))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	printerJobComplete := make(chan struct{})
	go func() {
		defer close(printerJobComplete)
		periodicPrinterJob(ctx, state, interval)
	}()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()
	slog.Info("calculator server listening", "address", listener.Addr().String())

	var serveErr error
	select {
	case <-ctx.Done():
		slog.Info("shutting down calculator server")
	case serveErr = <-serverErrors:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			slog.Error("server failed", "error", serveErr)
		}
	}
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		slog.Error("graceful shutdown failed", "error", shutdownErr)
		if err := server.Close(); err != nil {
			slog.Error("failed to close server", "error", err)
		}
	}

	<-printerJobComplete
	state.PrintTotals("final")
	if shutdownErr != nil || (serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed)) {
		os.Exit(1)
	}
}
