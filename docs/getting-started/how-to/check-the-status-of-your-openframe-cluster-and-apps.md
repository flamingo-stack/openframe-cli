# How to check the status of your OpenFrame cluster and apps

This task is for anyone operating an OpenFrame installation — a technician, administrator, or developer — who needs to confirm whether the cluster is reachable and whether the OpenFrame platform (ArgoCD and its applications) is up, synced, and healthy, using the `openframe` command-line tool.

## Prerequisites

Before you begin, confirm the following:

- The `openframe` CLI is installed and runs from your terminal (check with `openframe --help`).
- You have already created a cluster with `openframe cluster create` (or `openframe bootstrap`), or you have another reachable cluster you intend to check.
- Your kubectl context points at the cluster you want to inspect. You can check which cluster is active, or switch to one, with `openframe cluster use`.
- If you only want to see whether the cluster itself is healthy (nodes, resources) rather than the OpenFrame applications, use `openframe cluster status` instead of `openframe app status` — the two commands report different things.
- For `app status`, the OpenFrame platform (ArgoCD and the app-of-apps) should already be installed with `openframe app install`; the command still runs without it, but will tell you nothing is installed.

## Steps

1. Open a terminal and run `openframe app status` to get a one-time report of the platform: cluster reachability, the ArgoCD application table (name, sync, health), and a readiness summary.
2. Add `--context <name>` to `openframe app status` if you want to check a specific cluster instead of your current kube-context, for example `openframe app status --context k3d-openframe-dev`.
3. Add `--watch` (or `-w`) to `openframe app status` to keep the report on screen, refreshing automatically every few seconds, until you press Ctrl+C.
4. Add `--interactive` (or `-i`) to `openframe app status` to open the interactive, k9s-style view instead: use the Up/Down arrow keys (or `j`/`k`) to move between applications, press Enter to open an application's detail (revision, conditions, operation state) and Esc to go back, press `s` to trigger a sync of the selected application, press `r` to refresh immediately, and press `q` to quit.
5. Add `--output json` or `--output yaml` to `openframe app status` if you need the report in a machine-readable format instead of a table; this cannot be combined with `--watch` or `--interactive`.
6. Run `openframe cluster status` to check the underlying cluster itself — node status, resource usage, connectivity, and installed applications. Running it with no name opens an interactive cluster picker; run `openframe cluster status <NAME>` to target a specific cluster directly.
7. Add `--detailed` to `openframe cluster status` for more in-depth cluster information, or add `-o json` / `-o yaml` for the same information in machine-readable form.
8. Run `openframe app access` if the status report shows the platform is ready and you want the ArgoCD sign-in credentials and the port-forward command needed to reach the ArgoCD UI.

## Expected result

When you run `openframe app status`, you see:

- A reachability line, either `Cluster reachable (X/Y nodes ready)` in green, a yellow warning that node health could not be read (while the application table still appears), or a red `Cluster is not reachable. Is it running and is your kube-context correct?` if nothing could be read at all.
- A table headed `APPLICATION | SYNC | HEALTH` listing each ArgoCD application with its sync and health state.
- A one-line summary such as `X/Y synced, X/Y healthy — READY` printed in green when every application is synced and healthy, or the same line with `NOT READY` printed in yellow otherwise. If node health could not be read, the line adds `(node health unavailable)`.
- If no OpenFrame applications are found at all, a yellow warning: `No OpenFrame applications found — is it installed? Run: openframe app install`.
- If the platform is ready, access details (admin credentials and the port-forward command) printed right after the summary.

With `--watch`, the same information redraws in place every 3 seconds with a header line showing the current time, until you press Ctrl+C. With `--interactive`, you get the navigable application list with a footer reminding you of the keys: `↑/↓ navigate · enter detail · s sync · r refresh · q quit`.

`openframe cluster status` prints the cluster's health, node status, and installed applications as text by default, or as structured JSON/YAML when you pass `-o json` / `-o yaml`.

## Troubleshooting

- **`Cluster is not reachable. Is it running and is your kube-context correct?`** — the cluster could not be reached at all and no applications could be listed either. Confirm the cluster is running and that your kube-context is correct (use `openframe cluster use` to switch).
- **`Cluster node health could not be read: <cause>`** — the application list came back, meaning the API server answered, but reading node health failed (commonly an access/permissions issue on a namespace-scoped credential that cannot list cluster-wide nodes). The application table and summary are still shown; this is not the same as the cluster being down.
- **`No OpenFrame applications found — is it installed? Run: openframe app install`** — the cluster is reachable but the OpenFrame platform has not been installed yet, or was installed into a context the command cannot see. Install it following the install guide, then check status again.
- **`--watch/--interactive are live terminal views and cannot combine with --output <format>`** — remove `--watch` or `--interactive` when you want JSON or YAML output, or drop `--output` when you want the live view.
- **`--watch/--interactive need an interactive terminal (and cannot combine with --plain)`** — these views require running in a real interactive terminal; run the plain `openframe app status` (no flags) instead if you are in a non-interactive shell or script.
- **`"openframe app status" takes no positional arguments (got "...") — use --context to target a cluster`** — the command does not accept a cluster name as a plain argument; use `--context <name>` instead.
- **`invalid --output "<value>" (want "text", "json", or "yaml")`** from `openframe cluster status` — pass one of the three supported values to `-o`/`--output`.
- **`--output <format> requires a cluster name`** from `openframe cluster status` — machine-readable output never opens the interactive cluster picker; pass the cluster name as an argument when using `-o json` or `-o yaml`.
