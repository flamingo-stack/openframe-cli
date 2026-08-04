# Releasing

Versioning is fully automated by
[semantic-release](https://semantic-release.gitbook.io/) — there is no manual
version input anywhere. The version is computed from the **Conventional
Commits** on `main` since the last release tag:

| Commit subject | Effect |
|---|---|
| `fix: …`, `fix(scope): …` | patch bump |
| `feat: …` | minor bump |
| `feat!: …`, or a `BREAKING CHANGE:` footer | major bump |
| `chore:`, `docs:`, `refactor:`, `ci:`, `test:`, free-form | no release |

This repo squash-merges PRs, so the analyzed commit subjects are **PR titles**.
[pr-title.yml](../../.github/workflows/pr-title.yml) lints every PR title into
the conventional format — a free-form title would otherwise silently contribute
nothing to any release.

## Flow

Releases are **manual**: merging to `main` never ships anything by itself.
When you decide to ship, dispatch the workflow from `main`:

```bash
gh workflow run release.yml --ref main
```

1. [release.yml](../../.github/workflows/release.yml) then runs:
   - **plan** — `semantic-release --dry-run` resolves the next version from
     the commits since the last tag. No release-worthy commits → the run ends
     here as a no-op.
   - **release** (macOS runner) — lint + unit-test gates, signing setup, then
     `semantic-release` creates and pushes the bare `x.y.z` tag, and
     **GoReleaser** builds, signs and publishes the GitHub Release against it.
   - **verify** — the published Windows/macOS assets are downloaded on real
     runners and their signatures verified; a failure yanks the release + tag.
2. Re-dispatching after a rolled-back failure is safe (same rules re-apply).
   Dispatching from any other branch fails the guard step by design.

## Invariants (do not break)

- **Tags are bare `x.y.z`** (`tagFormat` in [.releaserc.yml](../../.releaserc.yml)):
  the self-updater and wsllauncher download
  `releases/download/<x.y.z>/…` URLs with no `v` prefix.
- **The workflow file must stay `.github/workflows/release.yml` and run from
  `main`**: the self-updater pins the cosign signing identity to
  `release.yml@refs/heads/main` (`internal/shared/selfupdate/cosign.go`).
- **GoReleaser owns the GitHub Release** — notes are label-based via
  [.github/release.yml](../../.github/release.yml) (`changelog.use:
  github-native`); semantic-release deliberately has no github/notes plugin.

Release *notes* (categorized, label-based, per PR) and release *versions*
(conventional-commit types) are thus decoupled: labels shape the changelog,
types shape the version.

Binary signing is documented in [release-signing.md](release-signing.md).
