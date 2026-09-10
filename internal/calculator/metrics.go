package calculator

import (
	"math"
	"slices"
	"sync"
	"time"
)

const (
	metricsWindow  = 60 * time.Second
	cLibraryKey    = "c"
	rustLibraryKey = "rust"
)

type Metrics struct {
	mutex    sync.Mutex
	requests [61]requestBucket
	calls    map[string]*callWindow
}

type requestBucket struct {
	second int64
	count  uint64
}

type callSample struct {
	finishedAt time.Time
	duration   time.Duration
}

type callWindow struct {
	samples []callSample
	head    int
}

type metricsSnapshot struct {
	rps              [60]uint64
	cP95, cP99       float64
	rustP95, rustP99 float64
}

func NewMetrics() *Metrics {
	return &Metrics{
		calls: map[string]*callWindow{
			cLibraryKey:    {},
			rustLibraryKey: {},
		},
	}
}

func (m *Metrics) recordRequest() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	second := now.Unix()
	bucket := &m.requests[second%int64(len(m.requests))]
	if bucket.second != second {
		*bucket = requestBucket{second: second}
	}
	bucket.count++
	m.expireCalls(now)
}

func (m *Metrics) recordCall(library string, finishedAt time.Time, duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	window := m.calls[library]
	window.samples = append(window.samples, callSample{finishedAt: finishedAt, duration: duration})
	m.expireCalls(time.Now())
}

func (m *Metrics) expireCalls(now time.Time) {
	cutoff := now.Add(-metricsWindow)
	for _, window := range m.calls {
		window.expire(cutoff)
	}
}

func (w *callWindow) expire(cutoff time.Time) {
	for w.head < len(w.samples) && w.samples[w.head].finishedAt.Before(cutoff) {
		w.samples[w.head] = callSample{}
		w.head++
	}
	if w.head == len(w.samples) {
		w.samples = nil
		w.head = 0
	} else if w.head >= len(w.samples)/2 && w.head > 0 {
		w.samples = slices.Clone(w.samples[w.head:])
		w.head = 0
	}
}

func (w *callWindow) durations() []time.Duration {
	durations := make([]time.Duration, len(w.samples)-w.head)
	for i, sample := range w.samples[w.head:] {
		durations[i] = sample.duration
	}
	return durations
}

func (m *Metrics) snapshot() metricsSnapshot {
	m.mutex.Lock()
	now := time.Now()
	m.expireCalls(now)

	var snapshot metricsSnapshot
	for i := range snapshot.rps {
		second := now.Unix() - int64(i+1)
		bucket := m.requests[second%int64(len(m.requests))]
		if bucket.second == second {
			snapshot.rps[i] = bucket.count
		}
	}
	cDurations := m.calls[cLibraryKey].durations()
	rustDurations := m.calls[rustLibraryKey].durations()
	m.mutex.Unlock()

	snapshot.cP95, snapshot.cP99 = durationQuantiles(cDurations)
	snapshot.rustP95, snapshot.rustP99 = durationQuantiles(rustDurations)
	return snapshot
}

func durationQuantiles(durations []time.Duration) (float64, float64) {
	if len(durations) == 0 {
		return math.NaN(), math.NaN()
	}
	slices.Sort(durations)
	p95 := int(math.Ceil(0.95*float64(len(durations)))) - 1
	p99 := int(math.Ceil(0.99*float64(len(durations)))) - 1
	return durations[p95].Seconds(), durations[p99].Seconds()
}
