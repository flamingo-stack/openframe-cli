#!/usr/bin/env bash
# Signs one GoReleaser-built binary before it is archived, so the published
# archives, checksums.txt and the cosign bundle all cover the signed bits.
# Invoked by the build post-hook in .goreleaser.yml:
#   sign-binary.sh <os> <arch> <path>
#
# No-op unless OPENFRAME_SIGN=1 (set only by the release workflow) — local
# builds and CI compile checks never attempt to sign.
#
# darwin  → codesign (Developer ID, hardened runtime) + notarytool, mirroring
#           .github/steps/sign-macos-package in openframe-oss-tenant.
# windows → Azure Trusted Signing (same AZURE_* secrets as openframe-oss-tenant)
#           via jsign, which unlike signtool runs on the macOS release runner.
# linux   → unsigned; covered by checksums.txt + cosign bundle only.
set -euo pipefail

OS="${1:?usage: sign-binary.sh <os> <arch> <path>}"
ARCH="${2:?usage: sign-binary.sh <os> <arch> <path>}"
BINARY="${3:?usage: sign-binary.sh <os> <arch> <path>}"

if [[ "${OPENFRAME_SIGN:-0}" != "1" ]]; then
  echo "sign-binary: OPENFRAME_SIGN != 1, skipping ${OS}/${ARCH}"
  exit 0
fi

sign_darwin() {
  : "${SIGNING_IDENTITY:?setup-macos-signing must run first}"
  : "${KEYCHAIN_PATH:?setup-macos-signing must run first}"
  : "${APPLE_ID_USERNAME:?}"
  : "${APPLE_ID_PASSWORD:?}"
  : "${APPLE_TEAM_ID:?}"

  codesign --sign "$SIGNING_IDENTITY" --keychain "$KEYCHAIN_PATH" \
    --timestamp --options runtime --force "$BINARY"
  codesign --verify --strict --verbose=2 "$BINARY"

  # Bare binaries can't be stapled; notarization is still recorded online and
  # checked by Gatekeeper on first run (same flow as openframe-oss-tenant).
  local zip
  zip="$(mktemp -d)/openframe-${OS}-${ARCH}.zip"
  zip -j "$zip" "$BINARY"
  xcrun notarytool submit "$zip" \
    --apple-id "$APPLE_ID_USERNAME" \
    --password "$APPLE_ID_PASSWORD" \
    --team-id "$APPLE_TEAM_ID" \
    --wait \
    --timeout 30m
  rm -f "$zip"

  echo "sign-binary: ${OS}/${ARCH} signed and notarized"
}

sign_windows() {
  : "${JSIGN_JAR:?install-jsign must run first}"
  : "${AZURE_TENANT_ID:?}"
  : "${AZURE_CLIENT_ID:?}"
  : "${AZURE_CLIENT_SECRET:?}"
  : "${AZURE_SIGNING_ENDPOINT:?}"
  : "${AZURE_CODE_SIGNING_ACCOUNT_NAME:?}"
  : "${AZURE_CERTIFICATE_PROFILE_NAME:?}"

  # Fetch the AAD token per binary: darwin notarization waits earlier in the
  # build can outlive a token fetched up-front.
  local token
  token="$(curl -fsS -X POST "https://login.microsoftonline.com/${AZURE_TENANT_ID}/oauth2/v2.0/token" \
    --data-urlencode "grant_type=client_credentials" \
    --data-urlencode "client_id=${AZURE_CLIENT_ID}" \
    --data-urlencode "client_secret=${AZURE_CLIENT_SECRET}" \
    --data-urlencode "scope=https://codesigning.azure.net/.default" \
    | jq -r '.access_token // empty')"
  if [[ -z "$token" ]]; then
    echo "sign-binary: failed to obtain Azure Trusted Signing token" >&2
    exit 1
  fi

  java -jar "$JSIGN_JAR" \
    --storetype TRUSTEDSIGNING \
    --keystore "${AZURE_SIGNING_ENDPOINT#https://}" \
    --storepass "$token" \
    --alias "${AZURE_CODE_SIGNING_ACCOUNT_NAME}/${AZURE_CERTIFICATE_PROFILE_NAME}" \
    --alg SHA-256 \
    --tsaurl http://timestamp.acs.microsoft.com \
    --tsmode RFC3161 \
    "$BINARY"

  echo "sign-binary: ${OS}/${ARCH} Authenticode-signed via Azure Trusted Signing"
}

case "$OS" in
  darwin)  sign_darwin ;;
  windows) sign_windows ;;
  *)       echo "sign-binary: ${OS}/${ARCH} not signed (by design)" ;;
esac
