package ui

import (
	"errors"
	"testing"
)

func TestShowOperationError(t *testing.T) {
	// Test that ShowOperationError doesn't panic with various inputs
	tips := []TroubleshootingTip{
		{Description: "Check status", Command: "kubectl get pods"},
		{Description: "Check logs", Command: "kubectl logs"},
	}

	// Should not panic with normal inputs
	ShowOperationError("test", "resource-name", errors.New("test error"), tips)

	// Should not panic with empty tips
	ShowOperationError("test", "resource-name", errors.New("test error"), []TroubleshootingTip{})

	// Should not panic with nil tips
	ShowOperationError("test", "resource-name", errors.New("test error"), nil)
}

func TestShowNoResourcesMessage(t *testing.T) {
	// Should not panic with normal inputs
	ShowNoResourcesMessage("clusters", "delete", "create command", "list command")

	// Should not panic with empty strings
	ShowNoResourcesMessage("", "", "", "")
}
