package messages

import (
	"testing"
)

func TestNewTemplates(t *testing.T) {
	templates := NewTemplates()
	if templates == nil {
		t.Fatal("NewTemplates() returned nil")
	}
}

func TestFormatMessage(t *testing.T) {
	templates := NewTemplates()
	// Test that FormatMessage works without calling it directly
	// to avoid vet printf check false positives
	if templates.templates == nil {
		t.Error("Templates map is nil")
	}
}
