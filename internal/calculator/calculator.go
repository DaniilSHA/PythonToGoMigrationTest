package calculator

import (
	"PythonToGoMigrationTest/internal/calculator/libraries"
	"PythonToGoMigrationTest/internal/config"
	"errors"
	"log/slog"
	"os"
	"time"
)

func Start() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.NewCalculatorConfig()
	if err != nil {
		return err
	}

	interval := time.Duration(cfg.Interval * float64(time.Second))
	if interval <= 0 {
		return errors.New("interval must be positive and fit within time.Duration")
	}

	loader := libraries.NewLibLoader(cfg)
	cLib, rustLib, err := loader.Load()

	if err != nil {
		return err
	}

	metrics := NewMetrics()
	state := NewState(cLib, rustLib, metrics)
	server := NewHTTPServer(cfg, state, metrics)
	return runServer(server, state, interval)
}
