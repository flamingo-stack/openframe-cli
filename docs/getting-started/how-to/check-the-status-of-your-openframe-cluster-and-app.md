
# How to check the status of your OpenFrame cluster and app

This task is for anyone operating an OpenFrame deployment — an administrator or technician — who needs to confirm that the Kubernetes cluster is reachable and that the OpenFrame platform (ArgoCD and its applications) is up and healthy.

## Prerequisites

- The `openframe` CLI is installed and on your `$PATH` (confirm by running `openframe --help`).
- You already have a cluster created with `openframe cluster create` (or `openframe bootstrap`), or any other reachable cluster you manage yourself.
- Your current kube-context points at the cluster you want to check, or you know the context name to pass with `--context`.
- The OpenFrame platform has been installed on the cluster with `openframe app install` (status can still be checked beforehand, but it will report that no applications are found).
- You are running the command from an interactive terminal if you intend to use the `--watch` or `--interactive` views — these do not work when combined with `--output` or `--plain`.

## Steps

1. Open a terminal and, if you need to target a specific cluster, confirm or switch your kube-context with `openframe cluster use [NAME]`.
2. Run `openframe app status` to get a one-time report of the platform's current state.
3. If you want to target a cluster other than your current kube-context without switching to it, add the `--context` flag, for example `openframe app status --context k3d-openframe-dev`.
4. If you want the view to keep refreshing automatically, add the `--watch` flag (short form `-w`): `openframe app status --watch`. The screen updates every few seconds; press `Ctrl+C` to stop watching.
5. If you want to navigate the list of applications, inspect one in detail, or trigger a sync from the same screen, add the `--interactive` flag (short form `-i`): `openframe app status --interactive`.
6. If you need the status in a script-friendly format instead of the table, add `--output json` or `--output yaml`, for example `openframe app status --output json`.
7. Read the printed report: the cluster reachability line, the applications table (APPLICATION, SYNC, HEALTH), and the one-line summary at the bottom.
8. If the summary shows the platform is ready and an admin password was found, note the ArgoCD access details that are printed along with it, or run `openframe app access` separately to print ArgoCD admin credentials and sign-in instructions.
9. While using the interactive view (step 5), use the on-screen keys to work the dashboard: arrow keys or `j`/`k` to move between applications, `Enter` to open an application's detail (revision, conditions, operation state) and `Esc` to close it, `s` to trigger a sync of the selected application, `r` to refresh immediately, and `q` or `Ctrl+C` to quit.

## Expected result

When the cluster is reachable and the platform is installed, a plain `openframe app status` prints:

- A success line: `Cluster reachable (N/M nodes ready)`.
- A table headed `APPLICATION  SYNC  HEALTH` listing every OpenFrame application with its sync and health state.
- A one-line summary such as `N/N synced, N/N healthy — READY` (or `— NOT READY` if any application is not yet synced or healthy).
- If an ArgoCD admin password could be read, the access details are printed right after the summary.

With `--watch`, the same information redraws in place every 3 seconds under a header line showing the current time and a `refresh 3s` / `Ctrl+C to exit` hint, until you press `Ctrl+C`.

With `--interactive`, you get a full-screen view headed `OpenFrame` with the one-line summary, the applications list (or the detail view for the selected application), and a footer reminding you of the navigation keys: `↑/↓ navigate · enter detail · s sync · r refresh · q quit`.

With `--output json` or `--output yaml`, you get a machine-readable document with fields for reachability, node counts, readiness, the summary line, and the full list of applications — suitable for piping into other tools instead of reading on screen.

## Troubleshooting

- **`Cluster is not reachable. Is it running and is your kube-context correct?`** — the cluster could not be reached at all and no applications were found either. Check that the cluster is running and that your kube-context (or the `--context` value you passed) is correct.
- **`Cluster node health could not be read`** (followed by a reason) — the applications list came back, meaning the API server answered, but reading node health failed, commonly because the credential in use is namespace-scoped and lacks permission to list nodes. The application table and summary are still shown; this is a permissions issue, not a down cluster.
- **`No OpenFrame applications found — is it installed? Run: openframe app install`** — the cluster is reachable but no applications exist yet. Run `openframe app install` to deploy the platform, then check status again.
- **`--watch/--interactive are live terminal views and cannot combine with --output <format>`** — remove `--output` when using `--watch` or `--interactive`, or drop `--watch`/`--interactive` when you need machine-readable output.
- **`--watch/--interactive need an interactive terminal (and cannot combine with --plain)`** — run the command directly in an interactive terminal session without `--plain`, or fall back to a plain `openframe app status` (optionally with `--output json`/`--output yaml`) in non-interactive contexts such as CI.
- **`"app status" takes no positional arguments (got "...") — use --context to target a cluster`** — you passed an extra word after `status`; remove it and use `--context <name>` instead to target a specific cluster.

The material available does not show how to check the status of a cluster itself (as opposed to the platform installed on it); that is a separate command, `openframe cluster status`, covered in its own documentation.
