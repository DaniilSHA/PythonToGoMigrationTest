package config

import (
	"flag"
	"os"
)

type GeneratorConfig struct {
	URL      string
	Threads  int
	Interval float64
	Timeout  float64
}

func NewGeneratorConfig() (*GeneratorConfig, error) {
	cfg := GeneratorConfig{}
	err := cfg.Init()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *GeneratorConfig) Init() error {
	parser := flag.NewFlagSet("Load generator for the calculator", flag.ContinueOnError)

	parser.StringVar(&cfg.URL, "url", "http://localhost:8080/calc", "calculator endpoint")
	parser.IntVar(&cfg.Threads, "threads", 10, "number of worker threads")
	parser.IntVar(&cfg.Threads, "n", 10, "number of worker threads")
	parser.Float64Var(&cfg.Interval, "interval", 0.1,
		"pause between requests per thread, in seconds (0 = as fast as possible)")
	parser.Float64Var(&cfg.Timeout, "timeout", 5.0, "HTTP request timeout, seconds")

	return parser.Parse(os.Args[1:])
}
