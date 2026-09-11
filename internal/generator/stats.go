package generator

import (
	"log/slog"
	"sync"
)

type Stats struct {
	mutex  sync.Mutex
	ok     uint64
	errors uint64
}

func (s *Stats) record(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err != nil {
		s.errors++
	} else {
		s.ok++
	}
}

func (s *Stats) PrintTotals() {
	s.mutex.Lock()
	ok, requestErrors := s.ok, s.errors
	s.mutex.Unlock()

	slog.Info("total requests", "ok", ok, "errors", requestErrors)
}
