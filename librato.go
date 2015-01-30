package metrics

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Sink provides means for submitting metrics to Librato service.
type Sink interface {
	// PostMetric submits a gauge or counter metric with given name and value
	// using HTTP POST.
	PostMetric(string, string, int64, time.Duration)
}

const (
	metricsEndpoint = "https://metrics-api.librato.com/v1/metrics"
	userAgent       = "IMTlibrato v1"
)

// NewSink returns Sink configured with given email, token and source.
func NewSink(email, token, hostname string, offline bool) Sink {
	return &librato{email: email, token: token, source: hostname, offline: offline}
}

type librato struct {
	email, token, source string
	offline              bool
}

func (l *librato) PostMetric(typ, name string, value int64, dur time.Duration) {
	if !l.offline {
		b := make(map[string][]map[string]interface{})
		if dur.Seconds() > 0 {
			b[typ] = []map[string]interface{}{{"name": name, "value": value, "source": l.source, "period": int64(dur.Seconds())}}
		} else {
			b[typ] = []map[string]interface{}{{"name": name, "value": value, "source": l.source}}
		}

		j, err := json.Marshal(b)
		if nil != err {
			log.Printf("Sink: Cannot marshal gauges %v: %v", b, err)
			return
		}
		go l.post(j)
	}
}

func (l *librato) post(body []byte) {
	req, err := http.NewRequest("POST", metricsEndpoint, bytes.NewBuffer(body))
	if nil != err {
		log.Printf("Sink: Cannot create metrics request: %v", err)
		return
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.SetBasicAuth(l.email, l.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Sink: Cannot post metrics request: %v", err)
		return
	}
	resp.Body.Close()
}
