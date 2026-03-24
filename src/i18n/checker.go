// i18n checker — validates translation completeness across protocols.
package i18n

import (
	"fmt"
	"strings"
)

// TranslationStatus represents the translation status of a protocol.
type TranslationStatus struct {
	Protocol     string   `json:"protocol"`
	Languages    []string `json:"languages"`
	Missing      []string `json:"missing"`
	Completeness float64  `json:"completeness"` // 0.0 to 1.0
}

// SupportedLanguages lists all supported languages.
var SupportedLanguages = []string{"en", "zh", "ja", "ko", "de", "fr", "es", "pt", "ru"}

// CheckTranslation checks translation completeness for a protocol's meta.
func CheckTranslation(title, description map[string]string) *TranslationStatus {
	status := &TranslationStatus{}

	for _, lang := range SupportedLanguages {
		hasTitle := title[lang] != ""
		hasDesc := description[lang] != ""
		if hasTitle && hasDesc {
			status.Languages = append(status.Languages, lang)
		} else if hasTitle || hasDesc {
			status.Languages = append(status.Languages, lang+"(partial)")
			status.Missing = append(status.Missing, lang)
		} else {
			status.Missing = append(status.Missing, lang)
		}
	}

	total := len(SupportedLanguages)
	if total > 0 {
		status.Completeness = float64(len(status.Languages)) / float64(total)
	}
	return status
}

// FormatTranslationReport formats translation status as a report.
func FormatTranslationReport(statuses []*TranslationStatus) string {
	var b strings.Builder
	b.WriteString("Translation Completeness Report\n")
	b.WriteString(strings.Repeat("=", 50) + "\n\n")

	for _, s := range statuses {
		icon := "✓"
		if s.Completeness < 0.5 {
			icon = "✗"
		} else if s.Completeness < 1.0 {
			icon = "~"
		}
		b.WriteString(fmt.Sprintf("%s %-20s %.0f%% (%d/%d languages)\n",
			icon, s.Protocol, s.Completeness*100,
			len(s.Languages), len(SupportedLanguages)))
		if len(s.Missing) > 0 {
			b.WriteString(fmt.Sprintf("  Missing: %s\n", strings.Join(s.Missing, ", ")))
		}
	}
	return b.String()
}
