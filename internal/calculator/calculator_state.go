package calculator

import (
	"fmt"
	"sync"
)

type State struct {
	cLib     *CLibrary
	rustLib  *RustLibrary
	mutex    sync.RWMutex
	sumValue int64
	subValue int64
}

func NewState(cLib *CLibrary, rustLib *RustLibrary) *State {
	return &State{cLib: cLib, rustLib: rustLib}
}

func (s *State) Add(num int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sumValue = s.cLib.Add(s.sumValue, num)
	s.subValue = s.rustLib.Sub(s.subValue, num)
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
