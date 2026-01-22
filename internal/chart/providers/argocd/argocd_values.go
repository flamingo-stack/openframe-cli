package argocd

// GetArgoCDValues returns the ArgoCD Helm chart values as YAML string
func GetArgoCDValues() string {
	return `fullnameOverride: argocd

global:
  image:
    repository: ghcr.io/flamingo-stack/registry/argoproj/argocd
    tag: v3.2.5

redis-ha:
  haproxy:
    image:
      repository: ghcr.io/flamingo-stack/registry/haproxy
  exporter:
    image:
      repository: ghcr.io/flamingo-stack/registry/oliver006/redis_exporter
      tag: v1.80.1

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
      repository: ghcr.io/flamingo-stack/registry/argoprojlabs/argocd-extension-installer
      tag: v0.0.9


repoServer:
  podAnnotations:
    loki.grafana.com/scrape: "true"
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 200m
      memory: 500Mi
  env:
    - name: ARGOCD_EXEC_TIMEOUT
      value: "180s"


redis:
  image:
    repository: ghcr.io/flamingo-stack/registry/redis
    tag: 8.2.2-alpine
  resources:
    requests:
      cpu: 10m
      memory: 32Mi
    limits:
      cpu: 50m
      memory: 32Mi


dex:
  image:
    repository: ghcr.io/flamingo-stack/registry/dexidp/dex
    tag: v2.44.0
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
`
}
