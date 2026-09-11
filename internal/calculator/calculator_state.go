package calculator

import (
	"PythonToGoMigrationTest/internal/calculator/libraries"
	"fmt"
	"sync"
	"time"
)

type State struct {
	cLib     *libraries.CLibrary
	rustLib  *libraries.RustLibrary
	metrics  *Metrics
	mutex    sync.RWMutex
	sumValue int64
	subValue int64
}

type nativeCallResult struct {
	value      int64
	finishedAt time.Time
	duration   time.Duration
}

func NewState(cLib *libraries.CLibrary, rustLib *libraries.RustLibrary, metrics *Metrics) *State {
	return &State{cLib: cLib, rustLib: rustLib, metrics: metrics}
}

func (s *State) Add(num int64) {
	cResultCh := make(chan nativeCallResult, 1)
	go func() {
		startedAt := time.Now()
		value := s.cLib.Add(0, num)
		finishedAt := time.Now()
		cResultCh <- nativeCallResult{
			value:      value,
			finishedAt: finishedAt,
			duration:   finishedAt.Sub(startedAt),
		}
	}()

	startedAt := time.Now()
	subDelta := s.rustLib.Sub(0, num)
	finishedAt := time.Now()
	cResult := <-cResultCh

	s.mutex.Lock()
	s.sumValue += cResult.value
	s.subValue += subDelta
	s.mutex.Unlock()

	s.metrics.recordCall(cLibraryKey, cResult.finishedAt, cResult.duration)
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
