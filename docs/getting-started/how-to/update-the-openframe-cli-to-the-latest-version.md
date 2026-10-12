# How to update the OpenFrame CLI to the latest version

This guide shows anyone running the `openframe` command-line tool how to check for a newer release and update to it, using the CLI's built-in `update` command.

## Prerequisites

- The `openframe` CLI is already installed and on your `$PATH` (running `openframe --version` prints a version instead of "command not found").
- You have an active internet connection, since updating downloads a release archive from GitHub.
- You are running the update from the same machine and account that owns the installed binary, so the CLI can overwrite its own executable file (on Linux and macOS this means the folder containing the binary must be writable by your user).
- If you access the CLI through GitHub Actions or any other API-limited environment, you have a `GITHUB_TOKEN` or `GH_TOKEN` environment variable available, since unauthenticated requests to GitHub are rate-limited and have been seen to fail checks with an HTTP 403.
- You are not running the native Windows launcher binary when expecting an in-place update — on Windows, self-update is not supported for the native launcher; see Troubleshooting below.

## Steps

1. Open a terminal and run `openframe update check` to see whether a newer release exists without changing anything yet.
2. Read the printed status: it reports either that you are already up to date, that a development build has no release to compare against, or that an update is available together with the current and latest version numbers and a release URL.
3. Run `openframe update` with no arguments to update to the latest published release.
4. When prompted, confirm the action at the `Update from <current> to <latest>?` prompt (the CLI shows "Update", "Downgrade", or "Reinstall" depending on the direction of the version change). Type the confirmation to proceed, or answer no to cancel.
5. If you want to skip the confirmation prompt (for example in a script), run `openframe update --yes` (or its short form, `openframe update -y`) instead of plain `openframe update`.
6. If you want to switch to a specific release instead of the latest one, run `openframe update v1.4.0`, replacing `v1.4.0` with the version you want — this works for moving to a newer or an older version.
7. If you want to force a reinstall of the version you already have, run `openframe update --force`.
8. Wait for the CLI to report its progress: it downloads the release, verifies the staged binary, and installs it in place.
9. Confirm the update finished by checking the final message, which names the new version the CLI is now running.

## Expected result

- `openframe update check` prints one of:
  - "Up to date (`<version>`)." when no update is available.
  - "Update available: `<current>` → `<latest>`" followed by the release URL, when a newer version exists.
  - "Development build — no release to compare against." when running a development build.
- `openframe update` without an available update prints "Already up to date (`<version>`)." and makes no changes.
- A successful update prints "OpenFrame CLI is now `<version>`." after the download, signature verification, and binary swap complete.
- If you cancel at the confirmation prompt, the CLI prints "Cancelled." and leaves the installed binary untouched.

## Troubleshooting

- **"This is a development build; self-update is disabled. Install a released version to enable updates."** — You are running a development build (not a tagged release). Install a released version of the CLI to enable the `update` command.
- **"self-update is not supported for the native Windows launcher; re-run the installer to update it (the WSL-side Linux binary updates itself)"** — On Windows, the native launcher binary cannot update itself in place. Re-run the Windows installer to get the latest launcher; if you installed through WSL, the Linux binary inside WSL updates itself normally.
- **"checking for updates: ..." with an HTTP error** — The release check failed, often because of GitHub API rate limiting on unauthenticated requests. Set a `GITHUB_TOKEN` or `GH_TOKEN` environment variable and try again.
- **"release `<version>` is not signed (no ..bundle), refusing to update; set OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY=1 to override at your own risk"** — The CLI could not find the expected signature bundle for that release and refuses to install an unverified binary. Only set `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY=1` if you understand this disables authenticity verification of the download; the material does not describe any other resolution.
- **"cannot update `<path>`: ... is not writable (installed with elevated privileges or by a package manager?)"** — The folder holding the installed binary is not writable by your user, typically because it was installed with elevated privileges or by a package manager. Re-run the update with the appropriate permissions for that install location.
- **"the new binary failed to run, keeping the current version: ..."** — The freshly downloaded binary failed a smoke test (`--version`), so the CLI automatically kept your current binary in place. No action is needed beyond retrying the update, since the original installation is preserved.
- **"installing the new binary (rolled back): ..."** — The swap to the new binary failed partway through, and the CLI restored your original binary automatically. Your existing installation remains usable; try the update again.
- **"could not save a rollback point: ... `openframe update rollback` will not work; the previous binary is kept at `<path>`"** — The update itself succeeded, but the CLI could not record a rollback point. The previous binary file is still on disk at the path shown if you need it, but `openframe update rollback` will not be able to restore it automatically.
- **Need to undo an update** — Run `openframe update rollback` to revert to the version saved by the last successful update, confirming at the prompt (or pass `--yes`/`-y` to skip it). If there is nothing to roll back to, it prints "No previous version to roll back to (nothing was saved by a prior update)."
