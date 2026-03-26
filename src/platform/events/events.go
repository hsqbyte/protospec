// Package events provides protocol event pub/sub and webhook notifications.
package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// EventType represents the type of protocol event.
type EventType string

const (
	EventProtocolChanged EventType = "protocol.changed"
	EventDecodeError     EventType = "decode.error"
	EventValidation      EventType = "validation.complete"
)

// Event represents a protocol event.
type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Protocol  string    `json:"protocol"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// FilterRule defines an event filter.
type FilterRule struct {
	Types     []EventType `json:"types,omitempty"`
	Protocols []string    `json:"protocols,omitempty"`
}

// Match checks if an event matches the filter.
func (f *FilterRule) Match(e *Event) bool {
	if len(f.Types) > 0 {
		found := false
		for _, t := range f.Types {
			if t == e.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(f.Protocols) > 0 {
		found := false
		for _, p := range f.Protocols {
			if p == e.Protocol {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Webhook represents a webhook endpoint.
type Webhook struct {
	URL    string     `json:"url"`
	Filter FilterRule `json:"filter"`
}

// EventBus manages event pub/sub.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]chan *Event
	webhooks    []Webhook
	history     []*Event
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]chan *Event),
	}
}

// Subscribe subscribes to events.
func (eb *EventBus) Subscribe(id string) chan *Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan *Event, 100)
	eb.subscribers[id] = ch
	return ch
}

// Unsubscribe removes a subscriber.
func (eb *EventBus) Unsubscribe(id string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if ch, ok := eb.subscribers[id]; ok {
		close(ch)
		delete(eb.subscribers, id)
	}
}

// Publish publishes an event.
func (eb *EventBus) Publish(e *Event) {
	eb.mu.Lock()
	eb.history = append(eb.history, e)
	eb.mu.Unlock()

	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, ch := range eb.subscribers {
		select {
		case ch <- e:
		default:
		}
	}
}

// AddWebhook registers a webhook.
func (eb *EventBus) AddWebhook(wh Webhook) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.webhooks = append(eb.webhooks, wh)
}

// Replay replays historical events.
func (eb *EventBus) Replay(filter *FilterRule) []*Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	var result []*Event
	for _, e := range eb.history {
		if filter == nil || filter.Match(e) {
			result = append(result, e)
		}
	}
	return result
}

// FormatEvent formats an event for display.
func FormatEvent(e *Event) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[%s] %s", e.Type, e.Protocol))
	if e.Data != nil {
		data, _ := json.Marshal(e.Data)
		b.WriteString(fmt.Sprintf(" — %s", string(data)))
	}
	return b.String()
}
