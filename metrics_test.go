package metrics

import (
	"testing"
	"time"
)

type fakeSink struct {
	email, token, source string
	typ, name            string
	value                int64
	dur                  time.Duration
}

func (l *fakeSink) PostMetric(typ, name string, value int64, dur time.Duration) {
	l.typ = typ
	l.name = name
	l.value = value
	l.dur = dur
}

func TestGauge(t *testing.T) {
	sink := &fakeSink{}
	m := &Service{sink: sink, metrics: make(map[string]*Metric)}
	g := m.Gauge("test1", time.Second*60)
	g.Set(9)
	if sink.typ != "gauges" {
		t.Errorf("sink.typ = %v; want gauges", sink.typ)
	}
	if sink.name != "test1" {
		t.Errorf("sink.name = %v; want test1", sink.name)
	}
	if sink.value != 9 {
		t.Errorf("sink.value = %v; want 9", sink.value)
	}
	if sink.dur.Seconds() != 60 {
		t.Errorf("sink.dur = %v; want 60", sink.dur)
	}
}

func TestCounter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	sink := &fakeSink{}
	m := &Service{sink: sink, metrics: make(map[string]*Metric)}
	c := m.Counter("test2", time.Second)
	c.Inc(90)
	time.Sleep(time.Second * 2)
	if sink.typ != "counters" {
		t.Errorf("sink.typ = %v; want counters", sink.typ)
	}
	if sink.name != "test2" {
		t.Errorf("sink.name = %v; want test2", sink.name)
	}
	if sink.value != 90 {
		t.Errorf("sink.value = %v; want 90", sink.value)
	}
	if sink.dur.Seconds() != 1 {
		t.Errorf("sink.dur = %v; want 1", sink.dur)
	}
}
