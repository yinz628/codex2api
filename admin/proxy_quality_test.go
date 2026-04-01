package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildProxyLocation(t *testing.T) {
	got := buildProxyLocation("United States", "California", "San Francisco")
	if got != "United States / California / San Francisco" {
		t.Fatalf("location = %q, want %q", got, "United States / California / San Francisco")
	}
}

func TestNormalizeProxyCountryFallsBackToLocation(t *testing.T) {
	got := normalizeProxyCountry("", "United States / California / San Francisco")
	if got != "United States" {
		t.Fatalf("country = %q, want %q", got, "United States")
	}
}

func TestRunProxyQualityCheckWarnsOnUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/v1/models")
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := server.Client()
	client.Transport = rewriteHostTransport{
		base: client.Transport,
		host: server.URL,
	}

	result := runProxyQualityCheck(client)
	if result.Status != proxyQualityStatusWarn {
		t.Fatalf("status = %q, want %q", result.Status, proxyQualityStatusWarn)
	}
	if result.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", result.StatusCode, http.StatusUnauthorized)
	}
}

func TestRunProxyQualityCheckFailsOnNonUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := server.Client()
	client.Transport = rewriteHostTransport{
		base: client.Transport,
		host: server.URL,
	}

	result := runProxyQualityCheck(client)
	if result.Status != proxyQualityStatusFail {
		t.Fatalf("status = %q, want %q", result.Status, proxyQualityStatusFail)
	}
	if result.StatusCode != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d", result.StatusCode, http.StatusForbidden)
	}
}

type rewriteHostTransport struct {
	base http.RoundTripper
	host string
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target := req.Clone(req.Context())
	target.URL.Scheme = "http"
	target.URL.Host = strings.TrimPrefix(t.host, "http://")
	target.Host = target.URL.Host
	return t.base.RoundTrip(target)
}
