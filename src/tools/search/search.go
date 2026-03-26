// Package search provides protocol documentation search engine.
package search

import (
	"fmt"
	"strings"
)

// Document represents a searchable protocol document.
type Document struct {
	Protocol string            `json:"protocol"`
	Fields   map[string]string `json:"fields"`
	Layer    string            `json:"layer"`
	Tags     []string          `json:"tags"`
}

// Index holds the search index.
type Index struct {
	Docs []Document
}

// NewIndex creates a new search index.
func NewIndex() *Index {
	return &Index{}
}

// Add adds a document to the index.
func (idx *Index) Add(doc Document) {
	idx.Docs = append(idx.Docs, doc)
}

// SearchResult represents a search result.
type SearchResult struct {
	Protocol string  `json:"protocol"`
	Score    float64 `json:"score"`
	Match    string  `json:"match"`
}

// Search performs a full-text search.
func (idx *Index) Search(query string) []SearchResult {
	query = strings.ToLower(query)
	var results []SearchResult
	for _, doc := range idx.Docs {
		score := 0.0
		match := ""
		if strings.Contains(strings.ToLower(doc.Protocol), query) {
			score += 10
			match = "protocol name"
		}
		for name, typ := range doc.Fields {
			if strings.Contains(strings.ToLower(name), query) {
				score += 5
				match = fmt.Sprintf("field: %s", name)
			}
			if strings.Contains(strings.ToLower(typ), query) {
				score += 2
				match = fmt.Sprintf("type: %s", typ)
			}
		}
		for _, tag := range doc.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				score += 3
				match = fmt.Sprintf("tag: %s", tag)
			}
		}
		if score > 0 {
			results = append(results, SearchResult{Protocol: doc.Protocol, Score: score, Match: match})
		}
	}
	return results
}

// FormatResults formats search results.
func FormatResults(results []SearchResult) string {
	var b strings.Builder
	for _, r := range results {
		b.WriteString(fmt.Sprintf("  %.0f  %s — %s\n", r.Score, r.Protocol, r.Match))
	}
	if len(results) == 0 {
		b.WriteString("  no results found\n")
	}
	return b.String()
}
