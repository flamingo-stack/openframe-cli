# How to update the OpenFrame CLI to the latest version

This guide shows a person who already has the `openframe` command installed how to check for a newer release and update it in place, including what to do if the update fails or needs to be undone.

## Prerequisites

- You can already run `openframe` from your terminal (any command, for example `openframe --version`, responds).
- You know which version you currently have installed — `openframe --version` prints it.
- You are not running on native Windows: self-update replaces the running executable in place, and the native Windows launcher does not support this. If you use OpenFrame CLI on Windows, you are expected to be working through the WSL-side Linux binary, which updates itself the normal way; re-running the installer is how the native Windows launcher itself gets updated.
- You have network access to GitHub, since the update downloads a release archive from this repository's releases.
- The folder that holds your `openframe` executable is writable by your user — installs placed there by a package manager or with elevated privileges can block the update (see Troubleshooting).

## Steps

1. Run `openframe update check` to see whether a newer release exists without changing anything on your machine.
2. Read the result: in the default text output it prints either "Up to date (`<current version>`)." or "Update available: `<current>` → `<latest>`" together with the release's URL.
3. Run `openframe update` with no arguments to update to the latest release.
4. Wait while the CLI checks for updates, then answer the confirmation prompt — it asks "Update from `<current>` to `<latest>`?" — by confirming, unless you already passed `--yes` to skip it.
5. Watch the progress messages while the new binary downloads and its signature and checksum are verified: "Downloading openframe `<version>`...", "Verifying the release signature (cosign / GitHub Actions identity)...", and "Verifying the staged binary...".
6. Confirm the final message "OpenFrame CLI is now `<version>`." appears, then run `openframe --version` to verify the installed version changed.
7. If you need a specific release instead of the latest one, run `openframe update <version>` (for example `openframe update v1.4.0`) — this works for moving to either a newer or an older release, and the confirmation prompt will say "Update", "Downgrade", or "Reinstall" depending on the direction.
8. If you want to update without being asked to confirm (for example in a script), add `--yes` to the command, and add `--force` if you want to reinstall even when you are already up to date.

## Expected result

- `openframe update check` with no newer release prints `Up to date (<current version>).` in green/success styling and makes no changes.
- `openframe update check` with a newer release available prints `Update available: <current> → <latest>` followed by the release's URL, and still makes no changes.
- A successful `openframe update` (or `openframe update <version>`) ends with the message `OpenFrame CLI is now <version>.` and `openframe --version` afterward reports the new version.
- If you decline the confirmation prompt, the CLI prints `Cancelled.` and leaves your installation untouched.
- If you run `openframe update check` with `--output json` or `--output yaml` (the `-o` flag on `update check`), the result is printed as structured JSON or YAML instead of the text messages above, which is useful for scripting.

## Troubleshooting

- **"This is a development build; self-update is disabled. Install a released version to enable updates."** — you are running a development build rather than a tagged release; the self-update mechanism has no released version to compare against. Install a released build to enable updates.
- **"self-update is not supported for the native Windows launcher; re-run the installer to update it (the WSL-side Linux binary updates itself)"** — you are on the native Windows executable. Re-run the Windows installer to get a newer launcher; the Linux binary it forwards into WSL2 updates itself normally.
- **"`<path>` is not writable (installed with elevated privileges or by a package manager?)"** — the folder containing your `openframe` executable cannot be written by your current user. The material does not show a supported way around this from inside the CLI; you will need to address the permissions or reinstall through whatever elevated or package-manager process placed the binary there.
- **"release `<version>` is not signed (no `<bundle file>`), refusing to update; set `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY=1` to override at your own risk"** — the CLI could not find the release's cosign signature bundle and refuses to install an unverifiable binary. Only set the escape-hatch environment variable if you understand this disables authenticity verification (checksums are still checked, but the download's origin is not).
- **"the new binary failed to run, keeping the current version"** — the freshly downloaded binary failed a quick self-check before being swapped in; the CLI keeps your current, working binary and reports this error instead of leaving you with a broken install.
- **Update appears to have broken something** — run `openframe update rollback` to revert to the previously installed version. If nothing was saved by a prior update, it prints `No previous version to roll back to (nothing was saved by a prior update).` and changes nothing. Otherwise it asks you to confirm with "Roll back from `<current>` to `<previous version>`?" (skip the prompt with `--yes`) and finishes with `Rolled back to <version>.`.
- **You want updates to happen automatically** — the material shows this is opt-in only, through the `OPENFRAME_AUTO_UPDATE` environment variable; it is checked at most once a day, it never runs in non-interactive or CI sessions, and it deliberately never applies a major-version bump automatically (it only prints a notice telling you to update manually in that case).

