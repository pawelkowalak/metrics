# metrics

[![GoDoc](https://godoc.org/github.com/viru/metrics?status.svg)](https://godoc.org/github.com/viru/metrics)

```
import "github.com/viru/metrics"

func main() {
	m := metrics.NewService(metrics.NewSink("email@address", "libratoToken", "hostname", false))

	// Increase child metric whenever we handle HTTP request successfuly.
	statusMetric := m.Counter("http.reqs", time.Minute)
	statusMetric.IncChild(200, 1)
```
