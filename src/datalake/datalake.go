// Package datalake provides protocol data lake integration.
package datalake

import (
	"fmt"
	"strings"
)

// ExportFormat represents a data export format.
type ExportFormat string

const (
	FormatParquet ExportFormat = "parquet"
	FormatArrow   ExportFormat = "arrow"
	FormatCSV     ExportFormat = "csv"
	FormatJSON    ExportFormat = "json"
)

// Column represents a data column derived from protocol fields.
type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// Schema represents a data lake schema for a protocol.
type Schema struct {
	Protocol string   `json:"protocol"`
	Columns  []Column `json:"columns"`
}

// DeriveSchema derives a data lake schema from protocol fields.
func DeriveSchema(protocol string, fields map[string]string) *Schema {
	s := &Schema{Protocol: protocol}
	for name, typ := range fields {
		colType := mapType(typ)
		s.Columns = append(s.Columns, Column{Name: name, Type: colType})
	}
	return s
}

func mapType(pslType string) string {
	switch {
	case strings.HasPrefix(pslType, "uint"):
		return "INT64"
	case strings.HasPrefix(pslType, "int"):
		return "INT64"
	case pslType == "bytes":
		return "BINARY"
	case pslType == "string":
		return "UTF8"
	case pslType == "bool":
		return "BOOLEAN"
	default:
		return "UTF8"
	}
}

// GenerateSQL generates a CREATE TABLE SQL statement.
func (s *Schema) GenerateSQL(tableName string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableName))
	for i, col := range s.Columns {
		sqlType := "TEXT"
		switch col.Type {
		case "INT64":
			sqlType = "BIGINT"
		case "BINARY":
			sqlType = "BLOB"
		case "BOOLEAN":
			sqlType = "BOOLEAN"
		}
		comma := ","
		if i == len(s.Columns)-1 {
			comma = ""
		}
		b.WriteString(fmt.Sprintf("  %s %s%s\n", col.Name, sqlType, comma))
	}
	b.WriteString(");\n")
	return b.String()
}

// GenerateDuckDBQuery generates a DuckDB query template.
func GenerateDuckDBQuery(table, protocol string) string {
	return fmt.Sprintf(`-- DuckDB query for %s protocol data
SELECT * FROM '%s.parquet'
WHERE protocol = '%s'
ORDER BY timestamp DESC
LIMIT 100;`, protocol, table, protocol)
}

// Describe returns a schema description.
func (s *Schema) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Schema for %s (%d columns):\n", s.Protocol, len(s.Columns)))
	for _, col := range s.Columns {
		b.WriteString(fmt.Sprintf("  %s: %s\n", col.Name, col.Type))
	}
	return b.String()
}
