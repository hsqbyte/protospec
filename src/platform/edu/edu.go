// Package edu provides interactive protocol education platform.
package edu

import (
	"fmt"
	"strings"
)

// Lesson represents a protocol tutorial lesson.
type Lesson struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Protocol string `json:"protocol"`
	Level    string `json:"level"` // beginner, intermediate, advanced
	Content  string `json:"content"`
}

// Quiz represents a knowledge quiz.
type Quiz struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   int      `json:"answer"`
}

// LearningPath represents a recommended learning path.
type LearningPath struct {
	Name    string   `json:"name"`
	Lessons []string `json:"lesson_ids"`
}

// Badge represents an achievement badge.
type Badge struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Criteria    string `json:"criteria"`
}

// Platform holds the education platform data.
type Platform struct {
	Lessons []Lesson
	Quizzes []Quiz
	Paths   []LearningPath
	Badges  []Badge
}

// NewPlatform creates a new education platform with default content.
func NewPlatform() *Platform {
	return &Platform{
		Lessons: []Lesson{
			{ID: "net-101", Title: "Network Basics", Protocol: "IPv4", Level: "beginner", Content: "Introduction to IPv4..."},
			{ID: "tcp-101", Title: "TCP Fundamentals", Protocol: "TCP", Level: "beginner", Content: "Understanding TCP..."},
			{ID: "dns-101", Title: "DNS Deep Dive", Protocol: "DNS", Level: "intermediate", Content: "DNS protocol analysis..."},
		},
		Paths: []LearningPath{
			{Name: "Network Fundamentals", Lessons: []string{"net-101", "tcp-101", "dns-101"}},
		},
		Badges: []Badge{
			{Name: "Protocol Explorer", Description: "Complete 5 lessons", Criteria: "lessons >= 5"},
			{Name: "Network Expert", Description: "Complete all network lessons", Criteria: "path:network complete"},
		},
	}
}

// ListLessons returns available lessons.
func (p *Platform) ListLessons() string {
	var b strings.Builder
	b.WriteString("Available Lessons:\n")
	for _, l := range p.Lessons {
		b.WriteString(fmt.Sprintf("  [%s] %s (%s) — %s\n", l.Level, l.Title, l.Protocol, l.ID))
	}
	return b.String()
}

// ListPaths returns learning paths.
func (p *Platform) ListPaths() string {
	var b strings.Builder
	b.WriteString("Learning Paths:\n")
	for _, path := range p.Paths {
		b.WriteString(fmt.Sprintf("  %s: %s\n", path.Name, strings.Join(path.Lessons, " → ")))
	}
	return b.String()
}

// ListBadges returns available badges.
func (p *Platform) ListBadges() string {
	var b strings.Builder
	b.WriteString("Badges:\n")
	for _, badge := range p.Badges {
		b.WriteString(fmt.Sprintf("  🏅 %s — %s\n", badge.Name, badge.Description))
	}
	return b.String()
}
