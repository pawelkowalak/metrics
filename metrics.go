// Package metrics provides simple service for sending metrics to Librato.
package metrics

import (
	"fmt"
	"sync"
	"time"
)

const (
	// counterPrefix is a counter prefix used by Librato.
	counterPrefix = "counters"
	// gaugePrefix is a gauge prefix used by Librato.
	gaugePrefix = "gauges"
)

// NewService returns Service configured with given sink.
func NewService(s Sink) *Service {
	return &Service{sink: s, metrics: make(map[string]*Metric)}
}

// Service allows clients to send gauges and aggregated counters using
// Librato sink.
type Service struct {
	sink    Sink
	mu      sync.Mutex
	metrics map[string]*Metric
}

// Counter returns a counter metric from a pool of existing metrics. If it
// doesn't exist, new metric is created and added to the pool first.
func (s *Service) Counter(name string, dur time.Duration) *Metric {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.metrics[fmt.Sprintf("%s%s", counterPrefix, name)]
	if m == nil {
		m = &Metric{s: s, name: name, typ: counterPrefix, dur: dur}
		s.metrics[fmt.Sprintf("%s%s", counterPrefix, name)] = m
	}
	return m
}

// Gauge returns a new gauge metric.
func (s *Service) Gauge(name string, dur time.Duration) *Metric {
	return &Metric{s: s, name: name, typ: gaugePrefix, dur: dur}
}

// PostMetric sends current value of the metric to Librato Sink.
func (s *Service) PostMetric(typ, name string, value int64, dur time.Duration) {
	s.sink.PostMetric(typ, name, value, dur)
}

// Metric is a base structure for all metrics.
type Metric struct {
	s      *Service
	typ    string
	name   string
	dur    time.Duration
	mu     sync.Mutex
	value  int64
	active bool
}

// Inc increments the counter metric value by given number. Flush job is started
// if the metric wasn't active before.
func (m *Metric) Inc(i int64) {
	if m.typ != counterPrefix {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value += i
	if !m.active {
		m.active = true
		go m.Flush()
	}
}

// IncChild increments the counter metric's child value by given number.
func (m *Metric) IncChild(child string, i int64) {
	ch := m.s.Counter(fmt.Sprintf("%s.%s", m.name, child), m.dur)
	ch.Inc(i)
}

// Set sets the gauge or counter metric value to given number.
func (m *Metric) Set(v int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.typ == gaugePrefix {
		m.s.PostMetric(gaugePrefix, m.name, v, m.dur)
	} else {
		m.s.PostMetric(counterPrefix, m.name, v-m.value, m.dur)
	}
	m.value = v
}

// Flush sends current value of the metric to Librato on given interval. Should
// be run as a goroutine.
func (m *Metric) Flush() {
	for _ = range time.Tick(m.dur) {
		m.mu.Lock()
		m.s.PostMetric(m.typ, m.name, m.value, m.dur)
		m.mu.Unlock()
	}
}

// MetricGroup holds a map of groupped metrics of different kinds and can be
// passed to periodic callback function for sending on given interval.
type MetricGroup struct {
	s       *Service
	dur     time.Duration
	mu      sync.Mutex
	metrics map[string]*Metric
}

func (g *MetricGroup) getMetric(prefix, name string) *Metric {
	g.mu.Lock()
	defer g.mu.Unlock()
	m := g.metrics[prefix+name]
	if m == nil {
		m = &Metric{s: g.s, name: name, typ: prefix, dur: g.dur}
		g.metrics[prefix+name] = m
	}
	return m
}

// Inc takes the counter metric with given name from the group and increments
// its value.
func (g *MetricGroup) Inc(name string, i int64) {
	m := g.getMetric(counterPrefix, name)
	m.Inc(i)
}

// Set takes the gauge metric with given name from the group and sets its value.
func (g *MetricGroup) Set(name string, i int64) {
	m := g.getMetric(gaugePrefix, name)
	m.Set(i)
}

// SetCounter takes the counter metric with given name from the group and
// sets its new value. In fact it increments the value by (new - old) amount.
func (g *MetricGroup) SetCounter(name string, i int64) {
	m := g.getMetric(counterPrefix, name)
	m.Set(i)
}

// NewPeriodicCallback triggers given callback function in a loop as a goroutine
// on given interval.
func (s *Service) NewPeriodicCallback(d time.Duration, cb func(MetricGroup)) {
	mg := MetricGroup{s: s, metrics: make(map[string]*Metric), dur: d}
	go func() {
		for _ = range time.Tick(d) {
			cb(mg)
		}
	}()
}
