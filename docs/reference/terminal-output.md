# Terminal output reference

The CLI picks an output mode per command from the terminal it runs in and the
flags it was given. Exactly one "live" surface is ever active at a time; every
other consumer gets sequential, log-friendly lines carrying the same
information.

## Output modes

| Mode | When | What you see |
|------|------|--------------|
| Live | interactive terminal, no `--verbose`/`--plain`/`--silent` | animated spinners, the in-place application dashboard, download progress bars |
| Sequential | redirected output, CI, `--plain`, `--verbose` | timestamped log lines: heartbeats with ready-deltas, stage lines, download begin/done announces |
| Silent | `--silent` | errors only |
| Machine | `-o json` / `-o yaml` | data on stdout, human warnings on stderr |

## Live surfaces

- **Stage checklist** (`bootstrap`) — each stage prints `◉ [2/3] Create
  cluster`, closes with `✔`/`✖` and its duration, and the run ends with a
  summary card: cluster + kube-context, per-stage timings, next commands.
- **Application dashboard** (install/bootstrap wait) — an in-place block with
  an animated header, elapsed time, a progress bar (`14/17 ready`), and the
  not-ready applications colored by health (red `Degraded`, yellow
  `Progressing`), capped at 8 with `+N more`. One-off events (stall hints,
  repo-server recovery notices) pin under the block as notes. The success
  line reports which applications took longest to become ready.
- **`app status --watch`** — the platform status re-rendered in place every
  3 s; a failing poll shows its error inside the view and keeps watching.
- **`app status --interactive`** — a k9s-style TUI over the ArgoCD
  applications: arrows/`j`/`k` navigate, `enter` opens the app detail (repo,
  path, target ref, revision, conditions, operation state), `s` triggers a
  per-app sync, `r` refreshes, `q` quits. Auto-refreshes every 3 s.
- **Download progress** — a self-rewriting line with a bar, percentage, and
  speed for verified tool downloads.
- **Desktop notification** — long operations (bootstrap, install) emit an
  OSC 9 notification plus a terminal bell on completion or failure, for the
  user who switched to another window.

`--watch` and `--interactive` require an interactive terminal and reject
machine output and `--plain`.

## Sequential mode

Where the live surfaces cannot run (redirected output, CI, `--plain`,
`--verbose`), the same information arrives as self-sufficient log lines:

- **Wait heartbeat** (every 30 s, 10 s under `--verbose`):
  `[12:34:05] apps 14/17 ready (+2 since last check) · elapsed 12m30s` with a
  `pending: tenant(Progressing), gateway(Degraded)` detail line. A `+0` delta
  makes a stall visible without diffing counts.
- **Interruption state** — Ctrl+C or a cancelled CI job during the wait
  records `interrupted at 14/17 applications ready · pending: …` before the
  cancellation error.
- **Download announces** — `Downloading helm-v3.16.2.tar.gz (52 MB)...` and
  `Downloaded … in 4.2s` replace the live bar.
- **Phase heartbeats** — output-less blocking operations (`helm --wait`)
  emit `<label> (2m30s elapsed)` every 30 s so a long phase is
  distinguishable from a hang.
- **Timestamped `--verbose`** — every debug line starts with a wall-clock
  stamp (`15:04:05.000`) so CLI actions can be correlated with cluster
  events.

## Failure panel

Failures render as an aligned panel instead of a wall of text:

```
✖ create failed for cluster big-gke
  cause   googleapi: Error 403: quota exceeded
  hint    💡 Request more quota, lower --nodes, or pick another region
  resume  openframe cluster create big-gke
```

`--verbose` adds the full error chain as a `chain` row.

## Flags and environment

| Control | Effect |
|---------|--------|
| `--plain` | sequential output with colors; no spinners, areas, or self-rewriting lines |
| `--silent` | errors only (also skips the logo and notifications) |
| `--verbose` | timestamped debug lines; disables the dashboard in favor of scrolling logs |
| `NO_COLOR` (non-empty) | strips all ANSI styling ([no-color.org](https://no-color.org)) |
| `CLICOLOR_FORCE=1` | keeps styling even when `NO_COLOR` is set |
| `OPENFRAME_ASCII=1` | plain ASCII glyphs (`ok`/`x`/`->`) instead of Unicode |
| `TERM=dumb` | implies ASCII glyphs and disables terminal escapes |
| `OPENFRAME_APP_WAIT_TIMEOUT` | application-wait budget (Go duration, default `60m`) |
| `OPENFRAME_DEGRADED_FAIL_AFTER` | how long an app may stay Degraded+Synced before the fail-fast considers it stuck (default `8m`) |

ANSI styling stays on for plain non-TTY output by default — modern CI log
renderers display it; `NO_COLOR` is the opt-out for consumers that don't.

Hyperlinks (OSC 8) are emitted only on terminals known to support them
(iTerm2, WezTerm, ghostty, Konsole, VTE-based, VS Code); everywhere else the
plain URL is printed.

## GitHub Actions

Inside a GitHub Actions job (`GITHUB_ACTIONS=true`):

- bootstrap stages fold into `::group::` log groups (the `✔`/`✖` closing
  lines stay outside the fold);
- the failure panel doubles as an `::error::` annotation with the headline
  and root cause, visible on the job and the PR;
- the closing bootstrap summary is appended to the job's Step Summary as a
  markdown table (stage, duration).
