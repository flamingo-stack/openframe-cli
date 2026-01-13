package argocd

// GetArgoCDValues returns the ArgoCD Helm chart values as YAML string
func GetArgoCDValues() string {
	return `fullnameOverride: argocd

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
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 1
      memory: 1Gi


server:
  podAnnotations:
    loki.grafana.com/scrape: "true"
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi


repoServer:
  podAnnotations:
    loki.grafana.com/scrape: "true"
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 768Mi
  env:
    - name: ARGOCD_EXEC_TIMEOUT
      value: "180s"


redis:
  resources:
    requests:
      cpu: 100m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 128Mi


dex:
  resources:
    requests:
      cpu: 10m
      memory: 32Mi
    limits:
      cpu: 50m
      memory: 64Mi


applicationSet:
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 100m
      memory: 128Mi


notifications:
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 100m
      memory: 128Mi
`
}
