package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMessage_KnownTemplate(t *testing.T) {
	tm := NewTemplates()
	got := tm.FormatMessage(InfoMessage, "operation_start", "deploy", "prod")
	assert.Equal(t, "📦 Starting deploy on cluster: prod", got)
}

func TestFormatMessage_SuccessTemplate(t *testing.T) {
	tm := NewTemplates()
	got := tm.FormatMessage(SuccessMessage, "operation_complete", "install")
	assert.Equal(t, "✅ install completed successfully!", got)
}

// The fallback path is the one that was reworked so `go vet`'s printf analyzer
// no longer mis-infers FormatMessage as a printf wrapper — the template key must
// NOT be treated as a format string.

func TestFormatMessage_UnknownKey_NoArgs(t *testing.T) {
	tm := NewTemplates()
	got := tm.FormatMessage(InfoMessage, "totally-unknown-key")
	assert.Equal(t, "totally-unknown-key", got)
}

func TestFormatMessage_UnknownKey_WithArgs(t *testing.T) {
	tm := NewTemplates()
	got := tm.FormatMessage(InfoMessage, "missing", "a", "b")
	assert.Equal(t, "missing: a b", got)
}

func TestFormatMessage_UnknownKeyWithPercentVerb_IsNotFormatted(t *testing.T) {
	tm := NewTemplates()
	// A key containing %s must be returned literally (not interpreted as a format
	// string), and args appended — proving the fallback never Sprintf's the key.
	got := tm.FormatMessage(InfoMessage, "weird %s key", "x")
	assert.Equal(t, "weird %s key: x", got)
	assert.NotContains(t, got, "%!", "must not produce a formatting error verb")
}

func TestFormatMessage_UnknownMessageType(t *testing.T) {
	tm := NewTemplates()
	got := tm.FormatMessage(MessageType(999), "operation_start")
	assert.Equal(t, "operation_start", got, "unknown message type falls back to the key")
}
