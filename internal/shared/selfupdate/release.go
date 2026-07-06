// Package selfupdate checks for newer published releases of the OpenFrame CLI
// and replaces the running binary in place.
//
// Phase 1 (this file set) verifies every artifact by SHA256 against the
// release's checksums.txt before it touches disk, reusing the verified-download
// substrate in internal/shared/download. Phase 2 will additionally verify the
// cosign (keyless) signature of checksums.txt with a pinned OIDC identity, so
// authenticity — not just integrity — is guaranteed.
package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// defaultAPIBase is the GitHub REST API root. Overridable per Client in tests.
	defaultAPIBase = "https://api.github.com"
	// repoOwner and repoName identify the release source (the module's own repo).
	repoOwner = "flamingo-stack"
	repoName  = "openframe-cli"
	// checksumsFile is the GoReleaser-published SHA256 listing covering every
	// release artifact.
	checksumsFile = "checksums.txt"
	// maxMetaBytes bounds the release-metadata and checksums reads.
	maxMetaBytes = 8 << 20 // 8 MiB
	userAgent    = "openframe-cli-selfupdate"
)

// Release is the subset of a GitHub release payload we consume.
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	HTMLURL    string  `json:"html_url"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Asset is one downloadable file attached to a release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// assetURL returns the download URL for the named asset.
func (r Release) assetURL(name string) (string, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a.URL, true
		}
	}
	return "", false
}

// Client fetches release metadata and checksums from GitHub. The zero value is
// usable and talks to the public API with a default HTTP client.
type Client struct {
	HTTP    *http.Client
	APIBase string // defaults to defaultAPIBase
	Token   string // optional; raises the unauthenticated rate limit
}

func (c Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (c Client) apiBase() string {
	if c.APIBase != "" {
		return c.APIBase
	}
	return defaultAPIBase
}

// Latest returns the repository's latest non-prerelease release. GitHub's
// /releases/latest endpoint already excludes drafts and prereleases, honouring
// each release's make_latest flag.
func (c Client) Latest(ctx context.Context) (Release, error) {
	return c.getRelease(ctx, "/repos/"+repoOwner+"/"+repoName+"/releases/latest")
}

// ForTag returns the release for an exact tag (e.g. "v1.2.3").
func (c Client) ForTag(ctx context.Context, tag string) (Release, error) {
	return c.getRelease(ctx, "/repos/"+repoOwner+"/"+repoName+"/releases/tags/"+tag)
}

func (c Client) getRelease(ctx context.Context, path string) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase()+path, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("querying releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("no matching release found")
	}
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("release query failed: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMetaBytes))
	if err != nil {
		return Release{}, fmt.Errorf("reading release metadata: %w", err)
	}
	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return Release{}, fmt.Errorf("decoding release metadata: %w", err)
	}
	return rel, nil
}

// fetchChecksum downloads the release's checksums.txt and returns the
// lowercase-hex SHA256 recorded for filename.
func (c Client) fetchChecksum(ctx context.Context, rel Release, filename string) (string, error) {
	url, ok := rel.assetURL(checksumsFile)
	if !ok {
		return "", fmt.Errorf("release %s has no %s", rel.TagName, checksumsFile)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", checksumsFile, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: HTTP %d", checksumsFile, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMetaBytes))
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", checksumsFile, err)
	}
	return parseChecksum(string(body), filename)
}

// parseChecksum extracts the hex digest for filename from a sha256sum listing
// ("<hex>␠␠<name>" per line; the name may carry a '*' binary-mode prefix).
func parseChecksum(listing, filename string) (string, error) {
	for _, line := range strings.Split(listing, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if strings.TrimPrefix(fields[1], "*") == filename {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("%s not listed in %s", filename, checksumsFile)
}
