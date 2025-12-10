package argocd

// GetArgoCDValues returns the ArgoCD Helm chart values as YAML string
func GetArgoCDValues() string {
	return `# global:
#   imagePullSecrets:
#     - name: docker-pat-secret

fullnameOverride: argocd

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
  replicas: 1
  resources:
    limits:
      cpu: "1"
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 256Mi
  env:
    - name: ARGOCD_EXEC_TIMEOUT
      value: "300s"
    - name: ARGOCD_GIT_ATTEMPTS_COUNT
      value: "5"
    - name: ARGOCD_GIT_RETRY_MAX_DURATION
      value: "30s"
  initContainers:
    - name: wait-for-dns
      image: busybox:1.36
      command: ['sh', '-c', 'until nslookup github.com; do echo waiting for DNS; sleep 2; done']
  # Increase parallelism limits to handle manifest generation
  extraArgs:
    - --parallelismlimit=2

redis:
  resources:
    limits:
      cpu: 100m
      memory: 128Mi
    requests:
      cpu: 50m
      memory: 64Mi
`
}

