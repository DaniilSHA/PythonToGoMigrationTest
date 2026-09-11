package libraries

import (
	"errors"
	"log/slog"

	"PythonToGoMigrationTest/internal/config"

	"github.com/ebitengine/purego"
)

type LibLoader struct {
	cfg *config.CalculatorConfig
}

func NewLibLoader(cfg *config.CalculatorConfig) *LibLoader {
	return &LibLoader{cfg: cfg}
}

type CLibrary struct {
	id  uintptr
	Add func(int64, int64) int64
}

type RustLibrary struct {
	id  uintptr
	Sub func(int64, int64) int64
}

func (ll LibLoader) Load() (*CLibrary, *RustLibrary, error) {
	if ll.cfg == nil {
		err := errors.New("calculator config is nil")
		slog.Error(err.Error())
		return nil, nil, err
	}

	id, add, err := loadLibFunction(ll.cfg.CLibPath, "add")
	if err != nil {
		return nil, nil, err
	}
	cLib := &CLibrary{id: id}
	purego.RegisterFunc(&cLib.Add, add)

	id, sub, err := loadLibFunction(ll.cfg.RustLibPath, "sub")
	if err != nil {
		return nil, nil, err
	}
	rustLib := &RustLibrary{id: id}
	purego.RegisterFunc(&rustLib.Sub, sub)

	return cLib, rustLib, nil
}

func loadLibFunction(path, name string) (uintptr, uintptr, error) {
	if path == "" {
		err := errors.New("library path is empty")
		slog.Error(err.Error(), "function", name)
		return 0, 0, err
	}
	id, err := openNativeLibrary(path)
	if err != nil {
		slog.Error("failed to load library", "path", path, "error", err)
		return 0, 0, err
	}
	function, err := findNativeFunction(id, name)
	if err != nil {
		_ = closeNativeLibrary(id)
		slog.Error("failed to find function", "path", path, "function", name, "error", err)
		return 0, 0, err
	}
	return id, function, nil
}
