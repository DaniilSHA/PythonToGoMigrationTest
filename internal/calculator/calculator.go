package calculator

import (
	"PythonToGoMigrationTest/internal/config"
	"log/slog"
	"os"
)

func Start() {
	cfg, err := config.NewCalculatorConfig()

	if err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}

	loader := NewLibLoader(cfg)
	cLib, rustLib, err := loader.load()

	if err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}

}
