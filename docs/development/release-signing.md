# Release Signing

Release binaries are code-signed during `goreleaser release` (see
[release.yml](../../.github/workflows/release.yml)), **before archiving** — so
the published archives, `checksums.txt`, and the cosign bundle all cover the
signed binaries.

| Platform | Mechanism |
|----------|-----------|
| macOS | `codesign` (Developer ID Application, hardened runtime, timestamp) + `notarytool` notarization |
| Windows | Authenticode via Azure Trusted Signing (SHA-256, RFC3161 timestamp from `timestamp.acs.microsoft.com`) |
| Linux | Unsigned; integrity via `checksums.txt` + cosign bundle |

The flow mirrors the `sign-macos-package` / `sign-windows-package` composite
steps in [openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
and uses the same certificates and secrets. One deviation: the release job runs
on a single macOS runner (signing must happen before GoReleaser packs the
archives), so Windows signing uses [jsign](https://ebourg.github.io/jsign/)
instead of `azure/trusted-signing-action` (signtool is Windows-only) — same
Azure Trusted Signing account, endpoint and certificate profile.

## How it's wired

- A GoReleaser build post-hook calls `scripts/sign-binary.sh <os> <arch> <path>`
  for every built binary. `linux` is a pass-through.
- The script is a no-op unless `OPENFRAME_SIGN=1`, which only the release
  workflow sets — local builds and CI compile checks never attempt to sign.
- [.github/steps/setup-macos-signing](../../.github/steps/setup-macos-signing/action.yml)
  imports the Developer ID certificate into a throwaway keychain and exports
  `KEYCHAIN_PATH` / `SIGNING_IDENTITY`; the keychain is deleted in an
  `if: always()` cleanup step.
- The workflow downloads a version-pinned, checksum-verified jsign jar and
  exports `JSIGN_JAR`. The script fetches a fresh AAD client-credentials token
  per Windows binary (notarization waits can outlive a token fetched up-front).

## Required secrets

Same names as `openframe-oss-tenant`, so org-level secrets cover both repos.

| Secret | Used for |
|--------|----------|
| `APPLE_CERTIFICATE_P12` | Base64-encoded Developer ID Application certificate (.p12) |
| `APPLE_CERTIFICATE_PASSWORD` | Password for the .p12 |
| `APPLE_ID_USERNAME` / `APPLE_ID_PASSWORD` | Notarization (app-specific password) |
| `APPLE_TEAM_ID` | Apple Developer Team ID |
| `AZURE_TENANT_ID` / `AZURE_CLIENT_ID` / `AZURE_CLIENT_SECRET` | AAD token for Trusted Signing |
| `AZURE_SIGNING_ENDPOINT` | e.g. `https://eus.codesigning.azure.net` |
| `AZURE_CODE_SIGNING_ACCOUNT_NAME` | Trusted Signing account |
| `AZURE_CERTIFICATE_PROFILE_NAME` | Certificate profile |

## Testing

Two layers, neither needing certificates locally:

- **Unit tests** — `tests/scripts/sign_binary_test.go` (part of
  `make test-unit`) runs `scripts/sign-binary.sh` with PATH stubs for
  `codesign`/`xcrun`/`java`/`curl`/`jq` that record their argv. They pin the
  `OPENFRAME_SIGN` gate, the per-OS dispatch, fail-fast on missing env, and the
  exact flags passed to codesign/notarytool/jsign (identity, hardened runtime,
  endpoint scheme-stripping, alias, timestamp URL, call ordering).
- **Post-publish verification** — the `verify-windows-signature` /
  `verify-macos-signature` jobs in the release workflow download the published
  assets on real Windows/macOS runners and verify them against the OS trust
  stores (`Get-AuthenticodeSignature` incl. timestamp; `codesign --verify` +
  Developer ID authority check, best-effort `spctl` notarization assessment).
  Both checks pin the signer identity — Authenticode subject
  `Flamingo AI, Inc.`, Apple `TeamIdentifier=F7LDSU8JPJ` — so a binary signed
  by *some* trusted-but-wrong publisher still fails. On verification failure
  the `cleanup-on-failed-verification` job yanks the release and tag.

## Verifying a released binary

macOS:

```bash
codesign --verify --strict --verbose=2 openframe
spctl --assess --type open --context context:primary-signature -v openframe
```

Windows (PowerShell):

```powershell
Get-AuthenticodeSignature .\openframe.exe
```

Any platform — release provenance (covers Linux too), see the `signs` block in
[.goreleaser.yml](../../.goreleaser.yml):

```bash
cosign verify-blob --bundle checksums.txt.bundle \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/flamingo-stack/openframe-cli/\.github/workflows/release\.yml@.*$' \
  checksums.txt
```
