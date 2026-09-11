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
	requestsMutex sync.Mutex
	requests      [61]requestBucket //храним 60 значения за 60 секунд + 1 текущее
	calls         map[string]*callWindow
}

type requestBucket struct {
	second int64 //метка времени
	count  uint64
}

type callSample struct {
	finishedAt time.Time
	duration   time.Duration
}

type callWindow struct {
	buckets [61]callBucket
}

type callBucket struct {
	mutex   sync.Mutex
	second  int64
	samples []callSample
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
	m.requestsMutex.Lock()
	defer m.requestsMutex.Unlock()

	now := time.Now()
	second := now.Unix()
	bucket := &m.requests[second%int64(len(m.requests))]
	if bucket.second != second {
		*bucket = requestBucket{second: second}
	}
	bucket.count++
}

func (m *Metrics) recordCall(libraryKey string, finishedAt time.Time, duration time.Duration) {
	window := m.calls[libraryKey]
	second := finishedAt.Unix()
	bucket := &window.buckets[second%int64(len(window.buckets))]
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	if finishedAt.Before(time.Now().Add(-metricsWindow)) || second < bucket.second {
		return
	}
	if bucket.second != second {
		bucket.second = second
		bucket.samples = nil
	}
	bucket.samples = append(bucket.samples, callSample{finishedAt: finishedAt, duration: duration})
}

func (w *callWindow) durations(now time.Time) []time.Duration {
	cutoff := now.Add(-metricsWindow)
	var durations []time.Duration
	for i := range w.buckets {
		values := w.buckets[i].durations(cutoff, now)
		durations = append(durations, values...)
	}
	return durations
}

func (b *callBucket) durations(cutoff, now time.Time) []time.Duration {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.second < cutoff.Unix() {
		b.samples = nil
		return nil
	}
	if b.second > now.Unix() {
		return nil
	}

	durations := make([]time.Duration, 0, len(b.samples))
	for _, sample := range b.samples {
		if !sample.finishedAt.Before(cutoff) && !sample.finishedAt.After(now) {
			durations = append(durations, sample.duration)
		}
	}
	return durations
}

func (m *Metrics) snapshot() metricsSnapshot {
	m.requestsMutex.Lock()
	now := time.Now()

	var snapshot metricsSnapshot
	for i := range snapshot.rps {
		second := now.Unix() - int64(i+1)
		bucket := m.requests[second%int64(len(m.requests))] //берем за последнии i секунд значение
		if bucket.second == second {
			snapshot.rps[i] = bucket.count
		}
	}
	m.requestsMutex.Unlock()
	cDurations := m.calls[cLibraryKey].durations(now)
	rustDurations := m.calls[rustLibraryKey].durations(now)

	snapshot.cP95, snapshot.cP99 = durationQuantiles(cDurations)
	snapshot.rustP95, snapshot.rustP99 = durationQuantiles(rustDurations)
	return snapshot
}

// используем метод nearest rank для расчетов перцентилей
// показывает max время за которое завершились N% вызовов функций
func durationQuantiles(durations []time.Duration) (float64, float64) {
	if len(durations) == 0 {
		return math.NaN(), math.NaN()
	}
	slices.Sort(durations)
	p95 := int(math.Ceil(0.95*float64(len(durations)))) - 1
	p99 := int(math.Ceil(0.99*float64(len(durations)))) - 1
	return durations[p95].Seconds(), durations[p99].Seconds()
}
