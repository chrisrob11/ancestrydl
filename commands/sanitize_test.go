package commands

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Portuguese characters
		{"João Silva", "Joao_Silva"},
		{"José María", "Jose_Maria"},
		{"São Paulo", "Sao_Paulo"},
		{"Conceição", "Conceicao"},
		{"Sebastião", "Sebastiao"},
		{"Antônio", "Antonio"},

		// Multiple accents
		{"Ação de Graças", "Acao_de_Gracas"},

		// Windows forbidden characters
		{"file/name", "file-name"},
		{"file\\name", "file-name"},
		{"file:name", "file-name"},
		{"file*name", "filename"},
		{"file?name", "filename"},
		{"file\"name", "filename"},
		{"file<name", "filename"},
		{"file>name", "filename"},
		{"file|name", "filename"},

		// Mixed Portuguese + forbidden chars
		{"João/Silva", "Joao-Silva"},
		{"José:María", "Jose-Maria"},

		// Empty and edge cases
		{"", "unnamed"},
		{"   ", "unnamed"},
		{"...", "unnamed"},

		// Long names (should be truncated to 50 chars)
		{"ThisIsAVeryLongNameThatExceedsFiftyCharactersAndShouldBeTruncated", "ThisIsAVeryLongNameThatExceedsFiftyCharactersAndSh"},

		// Trailing/leading special chars
		{"-João-", "Joao"},
		{"_María_", "Maria"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
