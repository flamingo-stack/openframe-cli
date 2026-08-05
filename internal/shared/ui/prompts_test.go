package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: These tests are limited because the prompts interact with a live
// terminal; interactive behavior is covered by PTY-driven integration testing.

// Test that the package exports the expected functions
func TestPackageExports(t *testing.T) {
	// Verify that all expected functions are available
	// This is more of a compile-time check, but ensures the API is stable

	t.Run("SelectFromList function exists", func(t *testing.T) {
		assert.NotNil(t, SelectFromList)
	})

	t.Run("SelectOption function exists", func(t *testing.T) {
		assert.NotNil(t, SelectOption)
	})

	t.Run("PromptInput function exists", func(t *testing.T) {
		assert.NotNil(t, PromptInput)
	})
}

// TestErrPromptInterrupted pins the exact error text: the shared error handler
// (errors.isInterruption) matches the literal "interrupted" to print a friendly
// cancellation notice, so a wording change here silently breaks Ctrl+C UX.
func TestErrPromptInterrupted(t *testing.T) {
	assert.Equal(t, "interrupted", ErrPromptInterrupted.Error())
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
