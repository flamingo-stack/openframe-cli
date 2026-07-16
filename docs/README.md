# OpenFrame CLI Documentation

`openframe` is an interactive command-line tool for standing up and managing OpenFrame Kubernetes environments. It provisions local k3d clusters, deploys the OpenFrame platform via an ArgoCD app-of-apps GitOps workflow, and keeps itself up to date.

This repository (`flamingo-stack/openframe-cli`) is the CLI. The platform and application manifests it deploys live in [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant).

## Getting Started

- [Introduction](./getting-started/introduction.md) — Overview and key concepts
- [Prerequisites](./getting-started/prerequisites.md) — System requirements and dependencies
- [Quick Start](./getting-started/quick-start.md) — Install and bootstrap in a few minutes
- [First Steps](./getting-started/first-steps.md) — Core commands and workflows

## Commands

- `openframe bootstrap` — Create a cluster and install the platform in one step
- `openframe cluster {create,delete,list,status,cleanup}` — Manage k3d clusters
- `openframe app {install,upgrade,status,access,uninstall}` — Manage the OpenFrame app-of-apps deployment
- `openframe prerequisites {check,install}` — Check and install required tools
- `openframe update` (`check`, `rollback`, `update <version>`) — Self-update the CLI
- `openframe completion` — Generate shell completion scripts

## System Requirements

A full local platform is demanding. Recommended host:

| Resource | Recommended |
|----------|-------------|
| RAM | 24 GB |
| CPU | 6 cores |
| Disk | 50 GB free |

## Dependencies

**Docker is the only tool you install and run yourself.** The CLI auto-installs pinned, verified copies of `kubectl`, `k3d`, and `helm` into `~/.openframe/bin`. `mkcert` is used to issue a locally-trusted certificate for the HTTPS ingress. See [Prerequisites](./getting-started/prerequisites.md).

## Community and Support

- **Slack**: [OpenMSP community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) (primary support channel)
- **Website**: [https://flamingo.run](https://flamingo.run)
- **Platform**: [https://openframe.ai](https://openframe.ai)

We don't monitor GitHub Issues for support — use Slack.

## License

See [LICENSE.md](../LICENSE.md).
