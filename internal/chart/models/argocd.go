package models

// ArgoCDImageConfig holds a single image configuration
type ArgoCDImageConfig struct {
	Repository string `json:"repository" yaml:"repository"`
	Tag        string `json:"tag,omitempty" yaml:"tag,omitempty"`
}

// ArgoCDConfig holds ArgoCD-specific configuration including images
type ArgoCDConfig struct {
	// Global ArgoCD image
	Image ArgoCDImageConfig `json:"image" yaml:"image"`
	// Redis image
	Redis ArgoCDImageConfig `json:"redis" yaml:"redis"`
	// Redis HA HAProxy image
	RedisHAProxy ArgoCDImageConfig `json:"redisHAProxy" yaml:"redisHAProxy"`
	// Redis exporter image
	RedisExporter ArgoCDImageConfig `json:"redisExporter" yaml:"redisExporter"`
	// Dex image
	Dex ArgoCDImageConfig `json:"dex" yaml:"dex"`
	// Extension installer image
	ExtensionInstaller ArgoCDImageConfig `json:"extensionInstaller" yaml:"extensionInstaller"`
}
