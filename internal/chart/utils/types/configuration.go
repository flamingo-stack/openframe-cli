package types

import (
	"time"

)

// DockerRegistryConfig holds Docker registry settings
type DockerRegistryConfig struct {
	Username string
	Password string
	Email    string
}

// DeploymentMode represents the deployment mode. The CLI supports only the
// OSS (oss-tenant) deployment; the type is retained as a single-valued enum.
type DeploymentMode string

const (
	DeploymentModeOSS DeploymentMode = "oss"
)

// IngressType represents the type of ingress to use
type IngressType string

const (
	IngressTypeLocalhost IngressType = "localhost"
	IngressTypeNgrok     IngressType = "ngrok"
	IngressTypeGCP       IngressType = "gcp"
)

// NgrokConfig holds Ngrok-specific configuration
type NgrokConfig struct {
	// Ngrok credentials
	AuthToken string `json:"authtoken"`
	APIKey    string `json:"apiKey"`
	Domain    string `json:"domain"`

	// IP allowlist configuration
	UseAllowedIPs bool     `json:"useAllowedIPs"`
	AllowedIPs    []string `json:"allowedIPs,omitempty"`

	// Registration tracking
	RegistrationCompleted bool      `json:"registrationCompleted,omitempty"`
	RegistrationStartTime time.Time `json:"registrationStartTime,omitempty"`
}

// IngressConfig holds ingress configuration options
type IngressConfig struct {
	Type        IngressType  `json:"type"`
	NgrokConfig *NgrokConfig `json:"ngrok,omitempty"`
}

// NgrokRegistrationURLs contains the URLs for Ngrok registration and documentation
var NgrokRegistrationURLs = struct { // #nosec G101 -- public ngrok documentation URLs, not credentials
	SignUp        string
	Dashboard     string
	APIKeyDocs    string
	AuthTokenDocs string
	DomainDocs    string
}{
	SignUp:        "https://dashboard.ngrok.com/signup",
	Dashboard:     "https://dashboard.ngrok.com",
	APIKeyDocs:    "https://dashboard.ngrok.com/api/new",
	AuthTokenDocs: "https://dashboard.ngrok.com/get-started/your-authtoken",
	DomainDocs:    "https://dashboard.ngrok.com/cloud-edge/domains",
}

// ChartConfiguration holds all configurable options for chart installation
type ChartConfiguration struct {
	BaseHelmValuesPath string                 // Path to the original openframe-helm-values.yaml (read-only)
	TempHelmValuesPath string                 // Path to the temporary helm values file for installation
	ExistingValues     map[string]interface{} // Current values from the file
	ModifiedSections   []string               // Track which sections were modified
	Branch             *string                // nil means use existing, otherwise use this value
	DockerRegistry     *DockerRegistryConfig  // nil means use existing, otherwise use this value
	IngressConfig      *IngressConfig         // nil means use existing, otherwise use this value
}

