package commands

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Transliterate Unicode characters to ASCII
	// This uses NFD (canonical decomposition) to separate base characters from accents,
	// then removes combining marks (accents, tildes, etc.), leaving just base characters
	// Example: ã (U+00E3) → a (U+0061) + ̃ (U+0303) → a (U+0061)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, name)
	if err != nil {
		// If transliteration fails, use original name
		result = name
	}
	name = result

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove or replace Windows-forbidden and problematic characters
	// Windows forbidden: \ / : * ? " < > |
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-", // Fixed: Double escape for literal backslash
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
		".", "",
	)
	name = replacer.Replace(name)

	// Remove any remaining non-ASCII characters as a safety measure
	var builder strings.Builder
	for _, r := range name {
		if r < 128 && (r == '_' || r == '-' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			builder.WriteRune(r)
		}
	}
	name = builder.String()

	// Limit length to avoid filesystem issues (Windows has 260-char path limit)
	if len(name) > 50 {
		name = name[:50]
	}

	// Trim any trailing hyphens or underscores
	name = strings.Trim(name, "-_ ")

	// Ensure the name is not empty
	if name == "" {
		name = "unnamed"
	}

	return name
}
