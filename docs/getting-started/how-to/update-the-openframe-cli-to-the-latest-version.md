# How to update the OpenFrame CLI to the latest version

derivative

# How to update the OpenFrame CLI to the latest version

This tutorial walks a person who already has the `openframe` command installed through checking whether a newer release exists and updating to it, for anyone running day-to-day OpenFrame CLI commands (`openframe cluster`, `openframe app`, `openframe bootstrap`, and so on) from their terminal.

## Prerequisites

- The `openframe` binary is already installed and runs from your terminal (confirm with `openframe --version`).
- You are on macOS or Linux. The native Windows build of the CLI cannot update itself in place — see Troubleshooting below.
- The folder containing the `openframe` binary is writable by your user account (installs done by a package manager or with elevated privileges may block the update — see Troubleshooting).
- You have network access to github.com, since the update downloads a release archive and its checksum/signature files from GitHub.
- Optional but recommended if you update often: a `GITHUB_TOKEN` or `GH_TOKEN` environment variable set, so the version check is not rate-limited.

## Steps

1. Open a terminal and run `openframe update check` to see whether a newer release is available without changing anything yet.
2. Read the result: a line reading "Up to date" means no action is needed; a line reading "Update available: `<current>` → `<latest>`" means a newer release exists.
3. Run `openframe update` with no arguments to update to the latest release (or `openframe update v1.4.0`, substituting the version you want, to switch to a specific release instead).
4. When prompted with "Update from `<current>` to `<latest>`?" (the prompt reads "Downgrade" or "Switch" instead of "Update" if you targeted an older or unparsable version), confirm by answering yes, or re-run the same command with the `--yes` flag to skip the prompt.
5. Wait while the CLI reports its progress: "Checking for updates...", then "Verifying the release signature (cosign / GitHub Actions identity)...", then "Downloading openframe `<version>`...", then "Verifying the staged binary...".
6. Confirm the update finished by reading the final line, which reads "OpenFrame CLI is now `<version>`.".
7. Verify the installed version by running `openframe --version` again.

## Expected result

- `openframe update check` prints either "Up to date (`<version>`)." or "Update available: `<current>` → `<latest>`" followed by a line with the release URL.
- A successful `openframe update` ends with the line "OpenFrame CLI is now `<version>`." and `openframe --version` afterward shows the new version number.
- If you answer "no" at the confirmation prompt, the CLI prints "Cancelled." and leaves the current binary untouched.
- If you decide you want the previous version back, running `openframe update rollback` (confirming the prompt the same way) ends with "Rolled back to `<version>`." — this works offline since it reuses the binary kept from the last update, not a fresh download.

## Troubleshooting

- **"self-update is not supported for the native Windows launcher; re-run the installer to update it (the WSL-side Linux binary updates itself)"** — this message means you are running the native Windows build of the CLI. The material does not show a documented in-place update path for that build beyond re-running the installer; the WSL2-side Linux binary (used when the CLI forwards into WSL) updates itself normally through the steps above.
- **"cannot update `<path>`: ... not writable (installed with elevated privileges or by a package manager?)"** — the CLI could not write to the folder holding its own binary. Reinstall the binary somewhere your user account owns, or re-run the update with the privileges your installation method requires.
- **"release `<version>` is not signed (no ...), refusing to update; set OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY=1 to override at your own risk"** — the CLI refuses to install a release it cannot cryptographically verify. Only set that environment variable if you understand it disables the authenticity check (the download's integrity is still checked, but not its origin).
- **"the new binary failed to run, keeping the current version"** — the downloaded binary failed a quick smoke test; the CLI automatically keeps your current, working binary in place, so you are not left without a working `openframe` command. Try the update again, and check your network connection if it recurs.
- **"No previous version to roll back to (nothing was saved by a prior update)."** — `openframe update rollback` has nothing to revert to because no earlier update has run yet (or the rollback point could not be saved, which is reported as its own warning at update time). There is no documented way to roll back further than the single previous version the CLI retains.
- **An update check or download returns an HTTP error from GitHub** — this is commonly rate limiting on unauthenticated requests. Set a `GITHUB_TOKEN` or `GH_TOKEN` environment variable (the same conventions used by GitHub Actions and the `gh` CLI) and try again.
