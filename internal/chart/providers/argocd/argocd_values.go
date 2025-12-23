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
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 1
      memory: 1Gi


server:
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi


# Disable non-essential components for lightweight installation (especially CI/k3d)
dex:
  enabled: false

notifications:
  enabled: false

applicationSet:
  enabled: false

# Resource constraints to prevent k3d/CI cluster overload
controller:
  resources:
    limits:
      cpu: "1"
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 512Mi
  env:
    - name: ARGOCD_RECONCILIATION_TIMEOUT
      value: "300s"
    - name: ARGOCD_REPO_SERVER_TIMEOUT_SECONDS
      value: "300"

server:
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 50m
      memory: 128Mi

repoServer:
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
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

