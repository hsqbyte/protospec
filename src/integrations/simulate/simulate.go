// Package simulate provides protocol simulation: mock servers, replay, and load testing.
package simulate

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// MockServer serves mock responses for a message protocol.
type MockServer struct {
	lib   *protocol.Library
	ms    *schema.MessageSchema
	rules map[string]json.RawMessage // custom response rules
	logs  []RequestLog
	mu    sync.Mutex
}

// RequestLog records a request to the mock server.
type RequestLog struct {
	Time     time.Time `json:"time"`
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	Body     string    `json:"body"`
	Response string    `json:"response"`
}

// NewMockServer creates a mock server for the given message protocol.
func NewMockServer(lib *protocol.Library, protoName string) (*MockServer, error) {
	ms := lib.Message(protoName)
	if ms == nil {
		return nil, fmt.Errorf("%s is not a message protocol", protoName)
	}
	return &MockServer{
		lib:   lib,
		ms:    ms,
		rules: make(map[string]json.RawMessage),
	}, nil
}

// LoadRules loads custom response rules from JSON.
func (m *MockServer) LoadRules(data []byte) error {
	return json.Unmarshal(data, &m.rules)
}

// Handler returns an http.Handler for the mock server.
func (m *MockServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handleRequest)
	mux.HandleFunc("/logs", m.handleLogs)
	return mux
}

func (m *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	w.Header().Set("Content-Type", "application/json")

	// Check custom rules
	if rule, ok := m.rules[path]; ok {
		resp := string(rule)
		m.logRequest(r, resp)
		fmt.Fprint(w, resp)
		return
	}

	// Find matching response message
	for _, msg := range m.ms.Messages {
		if msg.Kind == "response" && strings.EqualFold(msg.Name, path) {
			resp := generateMockResponse(msg.Fields)
			data, _ := json.MarshalIndent(resp, "", "  ")
			m.logRequest(r, string(data))
			w.Write(data)
			return
		}
	}

	// Auto-generate from any response
	for _, msg := range m.ms.Messages {
		if msg.Kind == "response" {
			resp := generateMockResponse(msg.Fields)
			data, _ := json.MarshalIndent(resp, "", "  ")
			m.logRequest(r, string(data))
			w.Write(data)
			return
		}
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, `{"error":"no matching response for %s"}`, path)
}

func (m *MockServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.logs)
}

func (m *MockServer) logRequest(r *http.Request, resp string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, RequestLog{
		Time:     time.Now(),
		Method:   r.Method,
		Path:     r.URL.Path,
		Response: resp,
	})
}

func generateMockResponse(fields []schema.MessageFieldDef) map[string]any {
	result := map[string]any{}
	for _, f := range fields {
		result[f.Name] = generateMockValue(f)
	}
	return result
}

func generateMockValue(f schema.MessageFieldDef) any {
	switch f.Type {
	case schema.MsgString:
		return fmt.Sprintf("mock_%s_%d", f.Name, rand.Intn(1000))
	case schema.MsgNumber:
		return rand.Intn(10000)
	case schema.MsgBoolean:
		return rand.Intn(2) == 1
	case schema.MsgObject:
		obj := map[string]any{}
		for _, sub := range f.Fields {
			obj[sub.Name] = generateMockValue(sub)
		}
		return obj
	case schema.MsgArray:
		arr := make([]any, rand.Intn(3)+1)
		for i := range arr {
			if f.ItemType != nil {
				arr[i] = generateMockValue(*f.ItemType)
			} else {
				arr[i] = fmt.Sprintf("item_%d", i)
			}
		}
		return arr
	default:
		return nil
	}
}

// --- Replay ---

// ReplayConfig holds replay configuration.
type ReplayConfig struct {
	Speed    float64 // Playback speed multiplier (1.0 = normal)
	Target   string  // Target address for replay
	Protocol string  // Protocol name
}

// ReplayPacket represents a packet to replay.
type ReplayPacket struct {
	Timestamp time.Time
	Data      []byte
	Protocol  string
}

// Replay replays packets with timing.
func Replay(packets []ReplayPacket, cfg *ReplayConfig) error {
	if len(packets) == 0 {
		return fmt.Errorf("no packets to replay")
	}

	conn, err := net.Dial("udp", cfg.Target)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", cfg.Target, err)
	}
	defer conn.Close()

	baseTime := packets[0].Timestamp
	for i, pkt := range packets {
		if i > 0 {
			delay := pkt.Timestamp.Sub(baseTime)
			if cfg.Speed > 0 {
				delay = time.Duration(float64(delay) / cfg.Speed)
			}
			elapsed := time.Since(packets[0].Timestamp)
			if delay > elapsed {
				time.Sleep(delay - elapsed)
			}
		}
		if _, err := conn.Write(pkt.Data); err != nil {
			return fmt.Errorf("send packet %d: %w", i, err)
		}
		fmt.Printf("replayed packet %d/%d (%d bytes)\n", i+1, len(packets), len(pkt.Data))
	}
	return nil
}

// --- Load Test ---

// LoadTestConfig holds load test configuration.
type LoadTestConfig struct {
	Target      string
	Protocol    string
	RPS         int // Requests per second
	Duration    time.Duration
	Concurrency int
	Data        []byte
}

// LoadTestResult holds load test results.
type LoadTestResult struct {
	TotalRequests int64         `json:"total_requests"`
	Successful    int64         `json:"successful"`
	Failed        int64         `json:"failed"`
	Duration      time.Duration `json:"duration"`
	RPS           float64       `json:"rps"`
	AvgLatency    time.Duration `json:"avg_latency"`
	P50Latency    time.Duration `json:"p50_latency"`
	P95Latency    time.Duration `json:"p95_latency"`
	P99Latency    time.Duration `json:"p99_latency"`
}

// RunLoadTest executes a protocol-level load test.
func RunLoadTest(cfg *LoadTestConfig) (*LoadTestResult, error) {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}
	if cfg.Duration <= 0 {
		cfg.Duration = 10 * time.Second
	}
	if cfg.RPS <= 0 {
		cfg.RPS = 100
	}

	var (
		total     int64
		success   int64
		failed    int64
		latencies []time.Duration
		mu        sync.Mutex
	)

	interval := time.Second / time.Duration(cfg.RPS)
	done := make(chan struct{})
	start := time.Now()

	// Worker pool
	work := make(chan struct{}, cfg.RPS*2)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range work {
				t0 := time.Now()
				conn, err := net.DialTimeout("udp", cfg.Target, 2*time.Second)
				if err != nil {
					atomic.AddInt64(&failed, 1)
					atomic.AddInt64(&total, 1)
					continue
				}
				_, err = conn.Write(cfg.Data)
				conn.Close()
				lat := time.Since(t0)

				atomic.AddInt64(&total, 1)
				if err != nil {
					atomic.AddInt64(&failed, 1)
				} else {
					atomic.AddInt64(&success, 1)
					mu.Lock()
					latencies = append(latencies, lat)
					mu.Unlock()
				}
			}
		}()
	}

	// Rate limiter
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				select {
				case work <- struct{}{}:
				default:
				}
			}
		}
	}()

	time.Sleep(cfg.Duration)
	close(done)
	close(work)
	wg.Wait()

	elapsed := time.Since(start)

	result := &LoadTestResult{
		TotalRequests: total,
		Successful:    success,
		Failed:        failed,
		Duration:      elapsed,
		RPS:           float64(total) / elapsed.Seconds(),
	}

	if len(latencies) > 0 {
		// Sort for percentiles
		sortDurations(latencies)
		result.AvgLatency = avgDuration(latencies)
		result.P50Latency = percentile(latencies, 50)
		result.P95Latency = percentile(latencies, 95)
		result.P99Latency = percentile(latencies, 99)
	}

	return result, nil
}

func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j] < d[j-1]; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}

func avgDuration(d []time.Duration) time.Duration {
	var sum time.Duration
	for _, v := range d {
		sum += v
	}
	return sum / time.Duration(len(d))
}

func percentile(d []time.Duration, p int) time.Duration {
	idx := len(d) * p / 100
	if idx >= len(d) {
		idx = len(d) - 1
	}
	return d[idx]
}
