package calculator

import (
	"PythonToGoMigrationTest/internal/calculator/handlers"
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

func NewHTTPServer(cfg *config.CalculatorConfig, state *State, metrics *Metrics) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("POST /calc", handlers.NewCalcHandler(state, metrics))
	mux.Handle("GET /metrics", handlers.NewMetricsHandler(metrics))
	mux.Handle("GET /health", handlers.NewHealthHandler())

	return &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler: mux,
	}
}

func runServer(server *http.Server, state *State, interval time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
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
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
	}
	stop()

	shutdownErr := shutdownServer(server)
	<-printerJobComplete
	state.PrintTotals("final")
	return errors.Join(serveErr, shutdownErr)
}

func shutdownServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return errors.Join(err, server.Close())
	}
	return nil
}
