// Package tenant provides multi-tenant protocol library isolation.
package tenant

import (
	"fmt"
	"strings"
	"sync"
)

// Tenant represents a tenant with isolated protocol library.
type Tenant struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Protocols []string `json:"protocols"`
	Quota     Quota    `json:"quota"`
}

// Quota represents tenant resource quotas.
type Quota struct {
	MaxProtocols int `json:"max_protocols"`
	MaxRequests  int `json:"max_requests_per_min"`
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	TenantID string `json:"tenant_id"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

// Manager manages multi-tenant protocol libraries.
type Manager struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	audit   []AuditEntry
}

// NewManager creates a new tenant manager.
func NewManager() *Manager {
	return &Manager{tenants: make(map[string]*Tenant)}
}

// CreateTenant creates a new tenant.
func (m *Manager) CreateTenant(id, name string, quota Quota) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[id]; ok {
		return fmt.Errorf("tenant already exists: %s", id)
	}
	m.tenants[id] = &Tenant{ID: id, Name: name, Quota: quota}
	m.audit = append(m.audit, AuditEntry{TenantID: id, Action: "create", Resource: "tenant"})
	return nil
}

// AddProtocol adds a protocol to a tenant.
func (m *Manager) AddProtocol(tenantID, protocol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tenants[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}
	if len(t.Protocols) >= t.Quota.MaxProtocols {
		return fmt.Errorf("quota exceeded for tenant %s", tenantID)
	}
	t.Protocols = append(t.Protocols, protocol)
	m.audit = append(m.audit, AuditEntry{TenantID: tenantID, Action: "add_protocol", Resource: protocol})
	return nil
}

// GetAuditLog returns audit entries for a tenant.
func (m *Manager) GetAuditLog(tenantID string) []AuditEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var entries []AuditEntry
	for _, e := range m.audit {
		if e.TenantID == tenantID {
			entries = append(entries, e)
		}
	}
	return entries
}

// Describe returns a description of all tenants.
func (m *Manager) Describe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Tenants (%d):\n", len(m.tenants)))
	for _, t := range m.tenants {
		b.WriteString(fmt.Sprintf("  %s (%s): %d protocols, quota=%d\n", t.Name, t.ID, len(t.Protocols), t.Quota.MaxProtocols))
	}
	return b.String()
}
