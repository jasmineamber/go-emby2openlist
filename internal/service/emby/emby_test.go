package emby

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRewriteEmbyWebSocketPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/socket?api_key=test", nil)
	rewriteEmbyWebSocketPath(req)

	if req.URL.Path != "/embywebsocket" {
		t.Fatalf("unexpected upstream path: %s", req.URL.Path)
	}
	if req.URL.RawQuery != "api_key=test" {
		t.Fatalf("query was changed: %s", req.URL.RawQuery)
	}
}

func TestRewriteEmbyWebSocketPathLeavesOtherPaths(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/other", nil)
	rewriteEmbyWebSocketPath(req)

	if req.URL.Path != "/other" {
		t.Fatalf("unexpected path rewrite: %s", req.URL.Path)
	}
}
