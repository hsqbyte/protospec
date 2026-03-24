// Package monitor provides real-time protocol traffic monitoring and alerting.
package monitor

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Stats holds real-time protocol statistics.
type Stats struct {
	mu            sync.RWMutex
	ProtocolCount map[string]*atomic.Int64
	TotalPackets  atomic.Int64
	TotalBytes    atomic.Int64
	StartTime     time.Time
	Sessions      map[string]*SessionInfo
}

// SessionInfo tracks a network session.
type SessionInfo struct {
	SrcAddr   string    `json:"src"`
	DstAddr   string    `json:"dst"`
	Protocol  string    `json:"protocol"`
	Packets   int64     `json:"packets"`
	Bytes     int64     `json:"bytes"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// NewStats creates a new Stats instance.
func NewStats() *Stats {
	return &Stats{
		ProtocolCount: make(map[string]*atomic.Int64),
		StartTime:     time.Now(),
		Sessions:      make(map[string]*SessionInfo),
	}
}

// Record records a decoded packet.
func (s *Stats) Record(proto string, size int, src, dst string) {
	s.TotalPackets.Add(1)
	s.TotalBytes.Add(int64(size))

	s.mu.Lock()
	if _, ok := s.ProtocolCount[proto]; !ok {
		s.ProtocolCount[proto] = &atomic.Int64{}
	}
	s.ProtocolCount[proto].Add(1)

	key := src + "->" + dst + ":" + proto
	if sess, ok := s.Sessions[key]; ok {
		sess.Packets++
		sess.Bytes += int64(size)
		sess.LastSeen = time.Now()
	} else {
		s.Sessions[key] = &SessionInfo{
			SrcAddr:   src,
			DstAddr:   dst,
			Protocol:  proto,
			Packets:   1,
			Bytes:     int64(size),
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
	}
	s.mu.Unlock()
}

// Snapshot returns a snapshot of current stats.
func (s *Stats) Snapshot() StatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elapsed := time.Since(s.StartTime).Seconds()
	total := s.TotalPackets.Load()
	totalBytes := s.TotalBytes.Load()

	snap := StatsSnapshot{
		TotalPackets: total,
		TotalBytes:   totalBytes,
		Duration:     time.Since(s.StartTime),
		PPS:          float64(total) / elapsed,
		BPS:          float64(totalBytes*8) / elapsed,
		Protocols:    make(map[string]int64),
	}

	for proto, cnt := range s.ProtocolCount {
		snap.Protocols[proto] = cnt.Load()
	}

	// Top sessions
	var sessions []*SessionInfo
	for _, sess := range s.Sessions {
		sessions = append(sessions, sess)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Bytes > sessions[j].Bytes
	})
	if len(sessions) > 10 {
		sessions = sessions[:10]
	}
	snap.TopSessions = sessions

	return snap
}

// StatsSnapshot is a point-in-time snapshot of stats.
type StatsSnapshot struct {
	TotalPackets int64            `json:"total_packets"`
	TotalBytes   int64            `json:"total_bytes"`
	Duration     time.Duration    `json:"duration"`
	PPS          float64          `json:"pps"`
	BPS          float64          `json:"bps"`
	Protocols    map[string]int64 `json:"protocols"`
	TopSessions  []*SessionInfo   `json:"top_sessions"`
}

// FormatText returns a text representation of the snapshot.
func (s StatsSnapshot) FormatText() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== Protocol Monitor (%s) ===\n", s.Duration.Truncate(time.Second)))
	b.WriteString(fmt.Sprintf("Packets: %d | Bytes: %d | PPS: %.1f | BPS: %.0f\n\n", s.TotalPackets, s.TotalBytes, s.PPS, s.BPS))

	b.WriteString("Protocol Distribution:\n")
	type kv struct {
		k string
		v int64
	}
	var sorted []kv
	for k, v := range s.Protocols {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	for _, p := range sorted {
		pct := float64(p.v) / float64(s.TotalPackets) * 100
		bar := strings.Repeat("█", int(pct/5))
		b.WriteString(fmt.Sprintf("  %-12s %6d (%5.1f%%) %s\n", p.k, p.v, pct, bar))
	}

	if len(s.TopSessions) > 0 {
		b.WriteString("\nTop Sessions:\n")
		for i, sess := range s.TopSessions {
			b.WriteString(fmt.Sprintf("  %d. %s → %s [%s] %d pkts, %d bytes\n",
				i+1, sess.SrcAddr, sess.DstAddr, sess.Protocol, sess.Packets, sess.Bytes))
		}
	}
	return b.String()
}

// --- Alert System ---

// AlertRule defines an alerting rule.
type AlertRule struct {
	Name      string `json:"name" yaml:"name"`
	Condition string `json:"condition" yaml:"condition"` // "field_match", "traffic_spike", "protocol_ratio"
	Protocol  string `json:"protocol,omitempty" yaml:"protocol"`
	Field     string `json:"field,omitempty" yaml:"field"`
	Operator  string `json:"operator,omitempty" yaml:"operator"` // ">", "<", "==", "!="
	Value     int64  `json:"value,omitempty" yaml:"value"`
	Action    string `json:"action" yaml:"action"` // "stdout", "webhook"
	Webhook   string `json:"webhook,omitempty" yaml:"webhook"`
}

// Alert represents a triggered alert.
type Alert struct {
	Rule      AlertRule `json:"rule"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// AlertEngine evaluates alert rules against stats.
type AlertEngine struct {
	rules  []AlertRule
	alerts []Alert
	mu     sync.Mutex
}

// NewAlertEngine creates a new alert engine.
func NewAlertEngine(rules []AlertRule) *AlertEngine {
	return &AlertEngine{rules: rules}
}

// LoadRules loads alert rules from JSON.
func (e *AlertEngine) LoadRules(data []byte) error {
	return json.Unmarshal(data, &e.rules)
}

// Evaluate checks all rules against current stats.
func (e *AlertEngine) Evaluate(snap StatsSnapshot) []Alert {
	var triggered []Alert
	for _, rule := range e.rules {
		if alert := e.evaluateRule(rule, snap); alert != nil {
			triggered = append(triggered, *alert)
		}
	}
	e.mu.Lock()
	e.alerts = append(e.alerts, triggered...)
	e.mu.Unlock()
	return triggered
}

func (e *AlertEngine) evaluateRule(rule AlertRule, snap StatsSnapshot) *Alert {
	switch rule.Condition {
	case "traffic_spike":
		if snap.PPS > float64(rule.Value) {
			return &Alert{
				Rule:      rule,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Traffic spike: %.0f pps > %d threshold", snap.PPS, rule.Value),
			}
		}
	case "protocol_ratio":
		if cnt, ok := snap.Protocols[rule.Protocol]; ok {
			ratio := float64(cnt) / float64(snap.TotalPackets) * 100
			if e.compare(int64(ratio), rule.Operator, rule.Value) {
				return &Alert{
					Rule:      rule,
					Timestamp: time.Now(),
					Message:   fmt.Sprintf("Protocol %s ratio: %.1f%% %s %d%%", rule.Protocol, ratio, rule.Operator, rule.Value),
				}
			}
		}
	}
	return nil
}

func (e *AlertEngine) compare(actual int64, op string, expected int64) bool {
	switch op {
	case ">":
		return actual > expected
	case "<":
		return actual < expected
	case ">=":
		return actual >= expected
	case "<=":
		return actual <= expected
	case "==":
		return actual == expected
	case "!=":
		return actual != expected
	default:
		return false
	}
}

// GetAlerts returns all triggered alerts.
func (e *AlertEngine) GetAlerts() []Alert {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]Alert{}, e.alerts...)
}
