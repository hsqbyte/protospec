// Package metrics provides protocol metrics collection and export.
package metrics

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Counter is a monotonically increasing counter.
type Counter struct {
	value atomic.Int64
}

func (c *Counter) Inc()         { c.value.Add(1) }
func (c *Counter) Add(n int64)  { c.value.Add(n) }
func (c *Counter) Value() int64 { return c.value.Load() }

// Histogram tracks value distributions.
type Histogram struct {
	mu     sync.Mutex
	values []float64
}

func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	h.values = append(h.values, v)
	h.mu.Unlock()
}

func (h *Histogram) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.values)
}

// Collector collects protocol metrics.
type Collector struct {
	DecodeTotal  Counter
	DecodeErrors Counter
	EncodeTotal  Counter
	DecodeLat    Histogram
	ProtoCount   map[string]*Counter
	mu           sync.Mutex
	startTime    time.Time
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{
		ProtoCount: make(map[string]*Counter),
		startTime:  time.Now(),
	}
}

// RecordDecode records a decode operation.
func (c *Collector) RecordDecode(proto string, elapsed time.Duration, err error) {
	c.DecodeTotal.Inc()
	c.DecodeLat.Observe(float64(elapsed.Microseconds()))
	if err != nil {
		c.DecodeErrors.Inc()
	}
	c.mu.Lock()
	if _, ok := c.ProtoCount[proto]; !ok {
		c.ProtoCount[proto] = &Counter{}
	}
	c.ProtoCount[proto].Inc()
	c.mu.Unlock()
}

// PrometheusExport exports metrics in Prometheus text format.
func (c *Collector) PrometheusExport() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# HELP psl_decode_total Total decode operations\n"))
	b.WriteString(fmt.Sprintf("# TYPE psl_decode_total counter\n"))
	b.WriteString(fmt.Sprintf("psl_decode_total %d\n", c.DecodeTotal.Value()))
	b.WriteString(fmt.Sprintf("# HELP psl_decode_errors_total Total decode errors\n"))
	b.WriteString(fmt.Sprintf("# TYPE psl_decode_errors_total counter\n"))
	b.WriteString(fmt.Sprintf("psl_decode_errors_total %d\n", c.DecodeErrors.Value()))
	b.WriteString(fmt.Sprintf("# HELP psl_decode_latency_count Decode latency sample count\n"))
	b.WriteString(fmt.Sprintf("# TYPE psl_decode_latency_count gauge\n"))
	b.WriteString(fmt.Sprintf("psl_decode_latency_count %d\n", c.DecodeLat.Count()))
	c.mu.Lock()
	for proto, cnt := range c.ProtoCount {
		b.WriteString(fmt.Sprintf("psl_protocol_decode_total{protocol=\"%s\"} %d\n", proto, cnt.Value()))
	}
	c.mu.Unlock()
	return b.String()
}

// GrafanaDashboard returns a Grafana dashboard JSON template.
func GrafanaDashboard() string {
	return `{
  "dashboard": {
    "title": "PSL Protocol Metrics",
    "panels": [
      {"title": "Decode Rate", "type": "graph", "targets": [{"expr": "rate(psl_decode_total[5m])"}]},
      {"title": "Error Rate", "type": "graph", "targets": [{"expr": "rate(psl_decode_errors_total[5m])"}]},
      {"title": "Protocol Distribution", "type": "piechart", "targets": [{"expr": "psl_protocol_decode_total"}]}
    ]
  }
}`
}
