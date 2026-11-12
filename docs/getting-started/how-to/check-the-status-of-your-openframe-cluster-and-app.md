# How to check the status of your OpenFrame cluster and app

A person running OpenFrame locally or in the cloud uses the `openframe` command-line tool to confirm the Kubernetes cluster is reachable and that the OpenFrame platform (ArgoCD and its applications) is deployed and healthy on it.

## Prerequisites

- The `openframe` CLI is installed and runs from your terminal (check with `openframe --help`).
- A cluster already exists, created with `openframe cluster create` or `openframe bootstrap` (see [Create a kubernetes cluster with OpenFrame CLI](./create-a-kubernetes-cluster-with-openframe-cli.md)), or another cluster you can already reach with `kubectl`.
- Your current kube-context points at the cluster you want to check, or you know the context name to pass with `--context`.
- If you intend to check the application status (not just the cluster), OpenFrame should already have been installed onto the cluster with `openframe app install` (see [Install the OpenFrame platform onto a cluster](./install-the-openframe-platform-onto-a-cluster.md)).
- Your terminal window is wide enough and interactive if you plan to use the `--watch` or `--interactive` view (these views refuse to run without an interactive terminal, and cannot be combined with the `--plain` flag).

## Steps

1. Open a terminal and run `openframe cluster status` to check the cluster itself. If you omit a cluster name, the CLI shows an interactive list of your managed clusters to pick from; alternatively run `openframe cluster status <NAME>` with the cluster's name.
2. Add `-o json` or `-o yaml` to `openframe cluster status <NAME>` if you need the cluster status in a machine-readable format (a cluster name is required for this; the interactive picker is skipped).
3. Run `openframe app status` to check whether the OpenFrame platform (ArgoCD and its applications) is up and running on the current cluster.
4. Add `--context <CONTEXT_NAME>` (or `-c <CONTEXT_NAME>`) to `openframe app status` if you want to check a cluster other than the one your current kube-context points to.
5. Add `--watch` (or `-w`) to `openframe app status` to keep the status on screen, refreshing automatically every few seconds; press Ctrl+C to stop watching.
6. Add `--interactive` (or `-i`) to `openframe app status` instead, to open an interactive view where you can navigate the list of applications, inspect details, and trigger a sync.
7. Add `-o json` or `-o yaml` to `openframe app status` if you need the platform status in a machine-readable format. This cannot be combined with `--watch` or `--interactive`.
8. Read the printed table of applications and the summary line at the bottom to judge overall readiness, as described under Expected result below.
9. If the status reports the platform is up (`READY`) and you want sign-in details for ArgoCD, run `openframe app access` to print the admin credentials and access instructions.

## Expected result

For `openframe cluster status`, you see the cluster's health, node status, and related information printed to the screen (or as JSON/YAML when `-o json`/`-o yaml` is used).

For `openframe app status` in the default text view, you see one of the following, depending on what is found:

- If the cluster cannot be reached at all and no applications were found, the message:
  `Cluster is not reachable. Is it running and is your kube-context correct?`
- If the application list did come back but node health could not be read (for example, an RBAC restriction), a warning:
  `Cluster node health could not be read`
- If the cluster is reachable, a success line such as:
  `Cluster reachable (N/M nodes ready)`

Then, if OpenFrame is installed, a table with columns `APPLICATION`, `SYNC`, and `HEALTH` listing every ArgoCD application, followed by a one-line summary such as `3/3 synced, 3/3 healthy — READY` (shown as a success line when ready, or as a warning with `NOT READY` when not every application is synced and healthy). If node health could not be read, the summary line additionally notes `(node health unavailable)`.

If no applications are found at all, you instead see the warning:
`No OpenFrame applications found — is it installed? Run: openframe app install`

When the platform is ready, the admin access information for ArgoCD is printed immediately after the summary line.

With `--watch`, the same information refreshes in place every 3 seconds, with a header line showing the current time and `refresh 3s`, until you press Ctrl+C.

With `-o json` or `-o yaml`, the output is a structured document instead of the table, containing the reachability, node counts, readiness, summary line, totals, and the list of applications with their names, sync state, and health state.

## Troubleshooting

- **`Cluster is not reachable. Is it running and is your kube-context correct?`** — The cluster could not be reached at all and no OpenFrame applications were found. Confirm the cluster is running and that your kube-context (or the `--context` value you passed) is correct.
- **`Cluster node health could not be read`** — The ArgoCD application list came back, meaning the API server answered, but reading node health failed (commonly because the credential in use lacks permission to list nodes cluster-wide). The application table and summary are still shown; this is a partial-information warning, not a failed status check.
- **`No OpenFrame applications found — is it installed? Run: openframe app install`** — The cluster is reachable but no OpenFrame applications exist on it yet. Install the platform with `openframe app install` before checking status again.
- **`--watch/--interactive are live terminal views and cannot combine with --output <format>`** — Remove `--watch` or `--interactive` if you need `-o json`/`-o yaml` output, or drop the output flag if you want the live view.
- **`--watch/--interactive need an interactive terminal (and cannot combine with --plain)`** — Run the command from an interactive terminal session without the `--plain` flag, or fall back to the plain `openframe app status` (without `--watch`/`--interactive`) when running in a non-interactive shell or script.
- **`"openframe app status" takes no positional arguments (got "...") — use --context to target a cluster`** — Do not pass a cluster name directly after `status`; use `--context <CONTEXT_NAME>` instead to target a specific cluster.
- **`invalid --output "<value>" (want "text", "json", or "yaml")`** (from `openframe cluster status`) — Use one of `text`, `json`, or `yaml` as the value for `-o`/`--output`.
- **`--output <format> requires a cluster name`** (from `openframe cluster status`) — Pass the cluster name as an argument, for example `openframe cluster status my-cluster -o json`, since machine-readable output cannot use the interactive cluster picker.
