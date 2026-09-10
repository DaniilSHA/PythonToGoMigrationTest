package config

import (
	"flag"
	"os"
	"path/filepath"
)

type CalculatorConfig struct {
	Host        string
	Port        int
	CLibPath    string
	RustLibPath string
	Interval    float64
}

func NewCalculatorConfig() (*CalculatorConfig, error) {
	cfg := CalculatorConfig{}
	err := cfg.Init()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *CalculatorConfig) Init() error {
	basePath, err := os.Executable()
	if err != nil {
		return err
	}
	scriptDir := filepath.Dir(basePath)

	parser := flag.NewFlagSet("Calculator HTTP server", flag.ContinueOnError)

	parser.StringVar(&cfg.Host, "host", "0.0.0.0", "")

	parser.IntVar(&cfg.Port, "port", 8080, "")

	parser.StringVar(&cfg.CLibPath, "c-lib", filepath.Join(scriptDir, "libcalculator.so"),
		"path to the compiled C shared library")

	parser.StringVar(&cfg.RustLibPath, "rust-lib", filepath.Join(scriptDir, "libcalculator_rust.so"),
		"path to the compiled Rust shared library")

	parser.Float64Var(&cfg.Interval, "interval", 5.0,
		"seconds between periodic sum/sub reports")

	return parser.Parse(os.Args[1:])
}
