# How to check the status of your OpenFrame cluster and app

This task shows a platform operator or developer how to confirm that an OpenFrame cluster is reachable and that the OpenFrame application installed on it (ArgoCD plus the app-of-apps) is synced and healthy, using the `openframe` CLI.

## Prerequisites

- The `openframe` CLI is installed and on your `$PATH`. Run `openframe --help` to confirm it runs.
- You already have a cluster created with `openframe cluster create` (or `openframe bootstrap`), or any other reachable Kubernetes cluster you have a kube-context for.
- Your kubectl context points at the cluster you want to check — `openframe app status` and `openframe cluster status` both default to the current kube-context when you don't pass one explicitly.
- If you only want cluster-level health (nodes, resources) without checking the OpenFrame application, you can skip straight to `openframe cluster status`; if you want to know whether the OpenFrame platform itself (ArgoCD applications) is up, use `openframe app status`.

## Steps

1. Open a terminal and confirm which cluster/context you're about to check by running `kubectl config current-context`, or note the cluster name you created with `openframe cluster create`.
2. Run `openframe cluster status` to see the Kubernetes-level health of a cluster. With no argument it opens an interactive cluster picker; pass a name directly, e.g. `openframe cluster status my-cluster`, to skip the picker.
3. Add `--detailed` to the cluster status command if you want deeper node/resource information (this flag is surfaced through the command's own flags).
4. Add `-o json` or `-o yaml` to `openframe cluster status my-cluster` if you want machine-readable output instead of the text view; machine output requires you to pass the cluster name explicitly (it will not fall back to the interactive picker).
5. Run `openframe app status` to check the OpenFrame platform itself: cluster reachability, the list of ArgoCD applications with their sync and health state, and an overall readiness summary.
6. Pass `--context <name>` to `openframe app status` if you want to check a cluster other than your current kube-context, for example `openframe app status --context k3d-openframe-dev`.
7. Add `--watch` (or `-w`) to `openframe app status` to keep the status on screen, refreshing automatically every few seconds; press `Ctrl+C` to stop watching.
8. Add `--interactive` (or `-i`) to `openframe app status` instead, to open the interactive status view where you can navigate the application list, inspect details, and trigger syncs.
9. Add `-o json` or `-o yaml` to `openframe app status` if you need the status as structured output for scripts; this cannot be combined with `--watch` or `--interactive`.
10. If the status reports a problem, follow the hint it prints — for example a "not installed yet" message suggests running `openframe app install`, which is covered in a separate task (see [Install the OpenFrame platform onto a cluster](./install-the-openframe-platform-onto-a-cluster.md)).

## Expected result

When `openframe app status` succeeds against a healthy platform, it prints:

- A success line: `Cluster reachable (<ready>/<total> nodes ready)`.
- A table with columns `APPLICATION`, `SYNC`, `HEALTH` listing every OpenFrame ArgoCD application and its state.
- A one-line summary such as `<synced>/<total> synced, <healthy>/<total> healthy — READY` (shown as a success line) when every application is synced and healthy, or the same line ending in `NOT READY` (shown as a warning) otherwise.
- If ArgoCD admin credentials are available, the status output also prints access details for signing in.

With `--watch`, the same information re-renders in place every 3 seconds with a header line showing the current time and a reminder that `Ctrl+C` exits.

With `-o json` or `-o yaml`, the output is a structured document with fields `reachable`, `nodesReady`, `nodesTotal`, `ready`, `summary`, `total`, `synced`, `healthy`, and an `applications` array of `{name, sync, health}` objects (plus `healthError` when node health could not be read).

`openframe cluster status` with no output flag prints cluster health, node status, installed applications, resource usage, and connectivity information as text; with `-o json` or `-o yaml` it prints the same `ClusterInfo` data as structured output.

## Troubleshooting

- **`Cluster is not reachable. Is it running and is your kube-context correct?`** — `openframe app status` could not reach the API server at all (no applications were listed either). Check that the cluster is running and that your kube-context is correct; the message is followed by a `cause` line showing the underlying error.
- **`Cluster node health could not be read: <error>`** — the application list came back (the API server answered) but reading node health failed, commonly because your credential is namespace-scoped and lacks cluster-wide permission to list nodes. The application table is still shown; this does not mean the platform is down.
- **`No OpenFrame applications found — is it installed? Run: openframe app install`** — the cluster is reachable but no OpenFrame applications exist yet. Install the platform first using the steps in [Install the OpenFrame platform onto a cluster](./install-the-openframe-platform-onto-a-cluster.md).
- **`--watch/--interactive are live terminal views and cannot combine with --output <format>`** — remove `-o`/`--output` when using `--watch` or `--interactive`, or drop `--watch`/`--interactive` when you need machine output.
- **`--watch/--interactive need an interactive terminal (and cannot combine with --plain)`** — these views require a real terminal session; run the command without `--plain` and outside of a non-interactive script/CI context, or use the plain `openframe app status` / `-o json` form instead.
- **`invalid --output "<value>" (want "text", "json", or "yaml")`** — from `openframe cluster status`, pass one of the three supported values to `-o`/`--output`.
- **`--output <format> requires a cluster name`** — when requesting `json` or `yaml` from `openframe cluster status`, you must pass the cluster name as an argument; it will not open the interactive picker for machine output.
- **`"<command>" takes no positional arguments (got "<value>") — use --context to target a cluster`** — `openframe app status` (and other `app` subcommands) reject stray positional arguments; use the `--context` flag to target a specific cluster instead of passing its name as an argument.

The material does not describe any other status-related failure messages beyond those shown above; if you see a different error, there is no documented explanation for it here.
