package calculator

import (
	"fmt"
	"sync"
	"time"
)

type State struct {
	cLib     *CLibrary
	rustLib  *RustLibrary
	metrics  *Metrics
	mutex    sync.RWMutex
	sumValue int64
	subValue int64
}

func NewState(cLib *CLibrary, rustLib *RustLibrary, metrics *Metrics) *State {
	return &State{cLib: cLib, rustLib: rustLib, metrics: metrics}
}

func (s *State) Add(num int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	startedAt := time.Now()
	sumValue := s.cLib.Add(s.sumValue, num)
	finishedAt := time.Now()
	s.sumValue = sumValue
	s.metrics.recordCall(cLibraryKey, finishedAt, finishedAt.Sub(startedAt))

	startedAt = time.Now()
	subValue := s.rustLib.Sub(s.subValue, num)
	finishedAt = time.Now()
	s.subValue = subValue
	s.metrics.recordCall(rustLibraryKey, finishedAt, finishedAt.Sub(startedAt))
}

func (s *State) Read() (int64, int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.sumValue, s.subValue
}

func (s *State) PrintTotals(label string) {
	sumValue, subValue := s.Read()
	fmt.Printf("[%s] sum=%d sub=%d\n", label, sumValue, subValue)
}
