package helm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var configMapGVR = schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}

func newFakeDynamic() *dynamicfake.FakeDynamicClient {
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{configMapGVR: "ConfigMapList"},
	)
}

const twoConfigMapsManifest = `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-one
  namespace: default
data:
  a: "1"
---
# a comment-only document must be skipped, not error
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-two
  namespace: default
data:
  b: "2"
`

func TestFetchManifest_Success(t *testing.T) {
	body := "hello: world\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	h := &HelmManager{manifestHTTPClient: srv.Client()}
	got, err := h.fetchManifest(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("fetchManifest: %v", err)
	}
	if string(got) != body {
		t.Fatalf("fetchManifest = %q, want %q", got, body)
	}
}

func TestFetchManifest_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	h := &HelmManager{manifestHTTPClient: srv.Client()}
	_, err := h.fetchManifest(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected a 404 status error, got %v", err)
	}
}

func TestFetchManifest_Timeout(t *testing.T) {
	// This reproduces (deterministically, no network) the failure mode the live
	// e2e hit: the manifest fetch exceeding the HTTP client timeout.
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-done // block until the test finishes
	}))
	defer srv.Close()
	defer close(done)

	client := srv.Client()
	client.Timeout = 50 * time.Millisecond
	h := &HelmManager{manifestHTTPClient: client}

	_, err := h.fetchManifest(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch manifest") {
		t.Fatalf("expected a fetch timeout error, got %v", err)
	}
}

func TestApplyManifestYAML_CreatesResources(t *testing.T) {
	dc := newFakeDynamic()
	h := &HelmManager{dynamicClient: dc}

	if err := h.applyManifestYAML(context.Background(), []byte(twoConfigMapsManifest)); err != nil {
		t.Fatalf("applyManifestYAML: %v", err)
	}

	for _, name := range []string{"cm-one", "cm-two"} {
		if _, err := dc.Resource(configMapGVR).Namespace("default").Get(context.Background(), name, metav1.GetOptions{}); err != nil {
			t.Errorf("expected %q to be created: %v", name, err)
		}
	}
}

func TestApplyManifestYAML_NilDynamicClient(t *testing.T) {
	h := &HelmManager{} // dynamicClient is nil
	err := h.applyManifestYAML(context.Background(), []byte(twoConfigMapsManifest))
	if err == nil || !strings.Contains(err.Error(), "dynamic client not initialized") {
		t.Fatalf("expected nil-client error, got %v", err)
	}
}

func TestApplyManifestYAML_MalformedYAML(t *testing.T) {
	h := &HelmManager{dynamicClient: newFakeDynamic()}
	err := h.applyManifestYAML(context.Background(), []byte("apiVersion: v1\nkind: ConfigMap\n  bad: : indent"))
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
}

func TestApplyManifestFromURL_EndToEnd(t *testing.T) {
	// The exact path the ArgoCD CRD install exercises: fetch a manifest over HTTP
	// and apply it — here with an httptest server and a fake dynamic client.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(twoConfigMapsManifest))
	}))
	defer srv.Close()

	dc := newFakeDynamic()
	h := &HelmManager{dynamicClient: dc, manifestHTTPClient: srv.Client()}

	if err := h.applyManifestFromURL(context.Background(), srv.URL); err != nil {
		t.Fatalf("applyManifestFromURL: %v", err)
	}
	if _, err := dc.Resource(configMapGVR).Namespace("default").Get(context.Background(), "cm-two", metav1.GetOptions{}); err != nil {
		t.Errorf("expected cm-two applied end-to-end: %v", err)
	}
}
