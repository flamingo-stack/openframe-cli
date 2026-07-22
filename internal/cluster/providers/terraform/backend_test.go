package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBackendURL(t *testing.T) {
	cases := []struct {
		in      string
		want    BackendConfig
		wantErr bool
	}{
		{"s3://bucket/some/prefix", BackendConfig{Scheme: "s3", Bucket: "bucket", Prefix: "some/prefix"}, false},
		{"gcs://bucket", BackendConfig{Scheme: "gcs", Bucket: "bucket"}, false},
		{"gcs://bucket/prefix", BackendConfig{Scheme: "gcs", Bucket: "bucket", Prefix: "prefix"}, false},
		{"http://bucket/prefix", BackendConfig{}, true},
		{"s3://", BackendConfig{}, true},
		{"just-a-bucket", BackendConfig{}, true},
		{`s3://bu"cket/prefix`, BackendConfig{}, true}, // HCL-hostile characters rejected
	}
	for _, tc := range cases {
		got, err := ParseBackendURL(tc.in)
		if tc.wantErr {
			assert.Error(t, err, "input %q", tc.in)
			continue
		}
		require.NoError(t, err, "input %q", tc.in)
		assert.Equal(t, tc.want, got, "input %q", tc.in)
	}
}

func TestWorkspace_WriteBackend(t *testing.T) {
	ws := OpenWorkspace(t.TempDir(), "demo")
	require.NoError(t, ws.WriteBackend([]byte("terraform {}\n")))

	data, err := os.ReadFile(filepath.Join(ws.TerraformDir(), "backend.tf"))
	require.NoError(t, err)
	assert.Equal(t, "terraform {}\n", string(data))
}
