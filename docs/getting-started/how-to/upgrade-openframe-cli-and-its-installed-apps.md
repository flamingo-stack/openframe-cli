# How to upgrade OpenFrame CLI and its installed apps

This task covers two related upgrades — updating the `openframe` command-line tool itself to a newer release, and upgrading the OpenFrame platform (ArgoCD and the app-of-apps) already deployed on a cluster — and is done by whoever operates the cluster, from a terminal with the `openframe` CLI installed.

## Prerequisites

- You have the `openframe` CLI installed and on your `$PATH`. Check with `openframe --version` or `openframe update check`.
- You have network access to GitHub (the CLI calls the GitHub Releases API to look up versions and download binaries).
- For a platform upgrade: you have an existing OpenFrame platform already installed on a reachable cluster with `openframe app install` — `openframe app upgrade` upgrades an **already-installed** platform, it does not install one.
- You know which kubectl context (or cluster name) the platform is installed on, since the upgrade commands can target a specific cluster.
- The directory containing the running `openframe` binary is writable by your user (an installation made by a package manager or an elevated/admin install may block the CLI's self-update — see Troubleshooting).
- On Windows, you are aware that the CLI's self-update (`openframe update`) runs only inside WSL2; the native Windows launcher cannot update itself this way.

## Steps

1. Check whether a new CLI release is available by running `openframe update check`. This only reports status and changes nothing.
2. Read the reported status: it prints either "Up to date (`<version>`)", "Update available: `<current>` → `<latest>`" with the release URL, or (on a development build) "Development build — no release to compare against."
3. Apply the latest CLI release by running `openframe update`. When a newer version exists you are prompted to confirm with a message naming the direction of the change ("Update", "Downgrade", or "Reinstall" from `<current>` to `<target>`); confirm with `y`, or pass `--yes` to skip the prompt (for example in a script).
4. To switch to one specific tagged release instead of the latest, run `openframe update v1.4.0` (replace `v1.4.0` with the desired tag). This works both to move up and to move back down to an older tag.
5. If you only want to reinstall the binary already on your machine (no version change), run `openframe update --force`.
6. If an update goes wrong and you need the previous binary back, run `openframe update rollback`. This reverts offline, with no download, and also asks for confirmation unless you pass `--yes`.
7. Once the CLI itself is current, upgrade the installed OpenFrame platform to pick up a new version of the app-of-apps: run `openframe app upgrade --ref v1.3.0` (replace `v1.3.0` with the release tag or branch you want to deploy). Add `--context <kube-context>` or pass a cluster name as an argument if the platform is not on your current kubectl context.
8. To preview a ref change before applying it, add `--dry-run` to the same command, for example `openframe app upgrade --ref main --dry-run`.
9. If you are not changing the deployed ref and only want ArgoCD to re-read git and re-sync the platform (for example after the branch it tracks moved forward), run `openframe app upgrade --sync` instead of `--ref`.
10. If the force-sync must also delete resources that were removed from git, add `--prune` to the sync command: `openframe app upgrade --sync --prune`. Only do this when you intend that deletion, since it is destructive.
11. Watch the command's output for the closing message confirming the result, then verify the platform's health afterward using `openframe app status` (see the separate how-to for checking cluster and app status).

## Expected result

- `openframe update check` prints one of: `Already up to date (<version>).`, `Update available: <current> → <latest>` followed by the release URL, or `Development build — no release to compare against.`
- A successful `openframe update` (or `openframe update <version>`) ends with `OpenFrame CLI is now <tag>.`
- A successful `openframe update rollback` ends with `Rolled back to <version-or-"the previous version">.`
- Running `openframe app upgrade --ref <tag>` prints `Upgrading OpenFrame to ref "<tag>"` and proceeds through the install sequence non-interactively, reusing your existing configuration.
- Running `openframe app upgrade --sync` (force-sync mode) prints `Refreshing and syncing the OpenFrame platform...` and finishes with `OpenFrame platform re-synced.`
- With `--sync --dry-run`, instead of applying anything you see a preview: either `All <N> applications are Synced; a force-sync would be a no-op.` or a count of applications that would be re-synced, listed as `OutOfSync: <name> (health=<health>, sync=<sync>)`.

## Troubleshooting

- **"`--ref` and `--sync` are mutually exclusive"** — you passed both flags to `openframe app upgrade`. Decide whether you want to deploy a new ref (`--ref`) or re-sync the current one (`--sync`), and drop the other flag.
- **"No ArgoCD applications found — is OpenFrame installed?"** (shown during `--sync --dry-run`) — the target cluster has no installed platform to upgrade; install it first rather than upgrading.
- **self-update refuses to run on the native Windows launcher** — `openframe update` is not supported for the native Windows binary; re-run the Windows installer to update it, since only the WSL-side Linux binary can self-update.
- **"cannot update `<path>`: ... not writable (installed with elevated privileges or by a package manager?)"** — the directory holding the running binary cannot be written to by your user; reinstall the CLI somewhere you own, or update it with the same elevated method you used to install it.
- **"the new binary failed to run, keeping the current version"** — the downloaded binary failed a smoke test, so the CLI automatically keeps your current version in place; try the update again.
- **"release `<tag>` is not signed (no ..bundle), refusing to update"** — the CLI could not verify the release's signature and will not proceed; this is a security check and should not be bypassed except knowingly, by setting the documented insecure-skip environment variable at your own risk.
- **"No previous version to roll back to (nothing was saved by a prior update)"** — `openframe update rollback` has nothing to revert to because no earlier update saved a backup; there is nothing further to do.
- **Update check fails with an HTTP error from GitHub** — this generally means the GitHub API rate limit was hit from your network; the material does not show a user-facing fix beyond retrying later.

For a related task, see how to check platform health after an upgrade in [Check the status of your OpenFrame cluster and apps](./check-the-status-of-your-openframe-cluster-and-apps.md), or how the platform gets installed in the first place in [Install the OpenFrame platform onto a cluster](./install-the-openframe-platform-onto-a-cluster.md).
