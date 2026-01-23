package argocd

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
)

// Default image configurations
var DefaultArgoCDImages = models.ArgoCDConfig{
	Image: models.ArgoCDImageConfig{
		Repository: "ghcr.io/flamingo-stack/registry/argoproj/argocd",
		Tag:        "v3.2.5",
	},
	Redis: models.ArgoCDImageConfig{
		Repository: "ghcr.io/flamingo-stack/registry/redis",
		Tag:        "8.2.2-alpine",
	},
	RedisHAProxy: models.ArgoCDImageConfig{
		Repository: "ghcr.io/flamingo-stack/registry/haproxy",
		Tag:        "",
	},
	RedisExporter: models.ArgoCDImageConfig{
		Repository: "ghcr.io/flamingo-stack/registry/oliver006/redis_exporter",
		Tag:        "v1.80.1",
	},
	Dex: models.ArgoCDImageConfig{
		Repository: "ghcr.io/flamingo-stack/registry/dexidp/dex",
		Tag:        "v2.44.0",
	},
	ExtensionInstaller: models.ArgoCDImageConfig{
		Repository: "ghcr.io/flamingo-stack/registry/argoprojlabs/argocd-extension-installer",
		Tag:        "v0.0.9",
	},
}

// GetArgoCDValues returns the ArgoCD Helm chart values as YAML string
// If config is nil, default images are used
func GetArgoCDValues(config *models.ArgoCDConfig) string {
	images := mergeArgoCDConfig(config)
	return generateArgoCDValuesYAML(images)
}

// mergeArgoCDConfig merges user config with defaults
func mergeArgoCDConfig(config *models.ArgoCDConfig) models.ArgoCDConfig {
	result := DefaultArgoCDImages

	if config == nil {
		return result
	}

	// Override only non-empty values
	if config.Image.Repository != "" {
		result.Image.Repository = config.Image.Repository
	}
	if config.Image.Tag != "" {
		result.Image.Tag = config.Image.Tag
	}

	if config.Redis.Repository != "" {
		result.Redis.Repository = config.Redis.Repository
	}
	if config.Redis.Tag != "" {
		result.Redis.Tag = config.Redis.Tag
	}

	if config.RedisHAProxy.Repository != "" {
		result.RedisHAProxy.Repository = config.RedisHAProxy.Repository
	}
	if config.RedisHAProxy.Tag != "" {
		result.RedisHAProxy.Tag = config.RedisHAProxy.Tag
	}

	if config.RedisExporter.Repository != "" {
		result.RedisExporter.Repository = config.RedisExporter.Repository
	}
	if config.RedisExporter.Tag != "" {
		result.RedisExporter.Tag = config.RedisExporter.Tag
	}

	if config.Dex.Repository != "" {
		result.Dex.Repository = config.Dex.Repository
	}
	if config.Dex.Tag != "" {
		result.Dex.Tag = config.Dex.Tag
	}

	if config.ExtensionInstaller.Repository != "" {
		result.ExtensionInstaller.Repository = config.ExtensionInstaller.Repository
	}
	if config.ExtensionInstaller.Tag != "" {
		result.ExtensionInstaller.Tag = config.ExtensionInstaller.Tag
	}

	return result
}

// generateArgoCDValuesYAML generates the ArgoCD values YAML from config
func generateArgoCDValuesYAML(images models.ArgoCDConfig) string {
	// Build redis-ha haproxy tag line only if tag is set
	haproxyTag := ""
	if images.RedisHAProxy.Tag != "" {
		haproxyTag = fmt.Sprintf("\n      tag: %s", images.RedisHAProxy.Tag)
	}

	return fmt.Sprintf(`fullnameOverride: argocd

global:
  image:
    repository: %s
    tag: %s

redis-ha:
  haproxy:
    image:
      repository: %s%s
  exporter:
    image:
      repository: %s
      tag: %s

configs:
  cm:
    resource.customizations.health.argoproj.io_Application: |
      hs = {}
      hs.status = "Progressing"
      hs.message = ""
      if obj.status ~= nil then
        if obj.status.health ~= nil then
          hs.status = obj.status.health.status
          if obj.status.health.message ~= nil then
            hs.message = obj.status.health.message
          end
        end
      end
      return hs
  params:
    controller.sync.timeout.seconds: "1800"


controller:
  podAnnotations:
    loki.grafana.com/scrape: "true"
  resources:
    requests:
      cpu: 100m
      memory: 600Mi
    limits:
      cpu: 200m
      memory: 800Mi


server:
  podAnnotations:
    loki.grafana.com/scrape: "true"
  resources:
    requests:
      cpu: 100m
      memory: 200Mi
    limits:
      cpu: 200m
      memory: 400Mi
  extensions:
    image:
      repository: %s
      tag: %s


repoServer:
  podAnnotations:
    loki.grafana.com/scrape: "true"
  resources:
    requests:
      cpu: 100m
      memory: 512Mi
    limits:
      cpu: 500m
      memory: 1Gi
  env:
    - name: ARGOCD_EXEC_TIMEOUT
      value: "180s"


redis:
  image:
    repository: %s
    tag: %s
  resources:
    requests:
      cpu: 10m
      memory: 32Mi
    limits:
      cpu: 50m
      memory: 32Mi


dex:
  image:
    repository: %s
    tag: %s
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
    limits:
      cpu: 50m
      memory: 64Mi


applicationSet:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
    limits:
      cpu: 50m
      memory: 64Mi


notifications:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
    limits:
      cpu: 50m
      memory: 64Mi
`,
		images.Image.Repository,
		images.Image.Tag,
		images.RedisHAProxy.Repository,
		haproxyTag,
		images.RedisExporter.Repository,
		images.RedisExporter.Tag,
		images.ExtensionInstaller.Repository,
		images.ExtensionInstaller.Tag,
		images.Redis.Repository,
		images.Redis.Tag,
		images.Dex.Repository,
		images.Dex.Tag,
	)
}
