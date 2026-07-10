package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: These tests are limited because promptui interacts with stdin/stdout
// In a real test environment, you would mock the promptui package or use integration tests

// TestSelectTemplates asserts on the shared selectTemplates the selectors use,
// so a styling change is a deliberate, reviewed edit.
func TestSelectTemplates(t *testing.T) {
	assert.Equal(t, "{{ . }}?", selectTemplates.Label)
	assert.Equal(t, "→ {{ . | cyan }}", selectTemplates.Active) // active row: arrow
	assert.Equal(t, "  {{ . | white }}", selectTemplates.Inactive)
	assert.Equal(t, "✓ {{ . | green }}", selectTemplates.Selected) // chosen row: check
}

// Test that the package exports the expected functions
func TestPackageExports(t *testing.T) {
	// Verify that all expected functions are available
	// This is more of a compile-time check, but ensures the API is stable

	t.Run("SelectFromList function exists", func(t *testing.T) {
		assert.NotNil(t, SelectFromList)
	})

}

func TestValidateNonEmpty(t *testing.T) {
	validator := ValidateNonEmpty("test field")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid input", "test", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"valid with spaces", "  test  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "test field cannot be empty")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	validator := ValidateIntRange(1, 10, "node count")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid number", "5", false},
		{"minimum valid", "1", false},
		{"maximum valid", "10", false},
		{"below minimum", "0", true},
		{"above maximum", "11", true},
		{"not a number", "abc", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
