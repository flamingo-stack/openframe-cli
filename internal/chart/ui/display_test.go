package ui

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/stretchr/testify/assert"
)

func TestNewDisplayService(t *testing.T) {
	service := NewDisplayService()
	assert.NotNil(t, service)
}

func TestChartTypeStrings(t *testing.T) {
	// Test that chart types can be converted to strings properly
	assert.Equal(t, "argocd", string(models.ChartTypeArgoCD))
}
