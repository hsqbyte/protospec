// Package ecosystem provides protocol ecosystem health dashboard.
package ecosystem

import (
	"fmt"
	"strings"
)

// HealthMetric represents an ecosystem health metric.
type HealthMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// ContribStats represents community contribution statistics.
type ContribStats struct {
	TotalProtocols int `json:"total_protocols"`
	Contributors   int `json:"contributors"`
	Downloads      int `json:"downloads"`
	Stars          int `json:"stars"`
}

// Trend represents a protocol usage trend.
type Trend struct {
	Protocol string  `json:"protocol"`
	Growth   float64 `json:"growth_percent"`
	Period   string  `json:"period"`
}

// Dashboard holds ecosystem dashboard data.
type Dashboard struct {
	Health  []HealthMetric `json:"health"`
	Contrib ContribStats   `json:"contributions"`
	Trends  []Trend        `json:"trends"`
}

// NewDashboard creates a dashboard with sample data.
func NewDashboard() *Dashboard {
	return &Dashboard{
		Health: []HealthMetric{
			{Name: "Protocol Coverage", Value: 85.5, Unit: "%"},
			{Name: "Test Pass Rate", Value: 98.2, Unit: "%"},
			{Name: "Doc Coverage", Value: 72.0, Unit: "%"},
		},
		Contrib: ContribStats{TotalProtocols: 50, Contributors: 120, Downloads: 50000, Stars: 2500},
		Trends: []Trend{
			{Protocol: "gRPC", Growth: 45.2, Period: "Q4"},
			{Protocol: "QUIC", Growth: 38.7, Period: "Q4"},
			{Protocol: "MQTT", Growth: 22.1, Period: "Q4"},
		},
	}
}

// Describe returns a text dashboard.
func (d *Dashboard) Describe() string {
	var b strings.Builder
	b.WriteString("=== PSL Ecosystem Dashboard ===\n\n")
	b.WriteString("Health Metrics:\n")
	for _, m := range d.Health {
		b.WriteString(fmt.Sprintf("  %s: %.1f%s\n", m.Name, m.Value, m.Unit))
	}
	b.WriteString(fmt.Sprintf("\nCommunity:\n  Protocols: %d | Contributors: %d | Downloads: %d | Stars: %d\n",
		d.Contrib.TotalProtocols, d.Contrib.Contributors, d.Contrib.Downloads, d.Contrib.Stars))
	b.WriteString("\nTrending Protocols:\n")
	for _, t := range d.Trends {
		b.WriteString(fmt.Sprintf("  📈 %s +%.1f%% (%s)\n", t.Protocol, t.Growth, t.Period))
	}
	return b.String()
}
