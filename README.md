# metrics

[![GoDoc](https://godoc.org/github.com/viru/metrics?status.svg)](https://godoc.org/github.com/viru/metrics) [![Build Status](https://semaphoreci.com/api/v1/projects/635ed6cc-e538-455c-b0cb-79e9de30daa9/481794/badge.svg)](https://semaphoreci.com/viru/metrics)

```
import "github.com/viru/metrics"

func main() {
	m := metrics.NewService(metrics.NewSink("email@address", "libratoToken", "hostname", false))

	// Increase child metric whenever we handle HTTP request successfuly.
	statusMetric := m.Counter("http.reqs", time.Minute)
	statusMetric.IncChild(200, 1)
```
