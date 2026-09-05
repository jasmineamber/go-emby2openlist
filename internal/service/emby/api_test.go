package emby

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
)

func TestBuildPlaybackInfoRequestHeaderPreservesAuthorization(t *testing.T) {
	itemInfo := ItemInfo{ApiKeyType: Query, ApiKey: "query-token"}
	got := buildPlaybackInfoRequestHeader(itemInfo, http.Header{
		"Authorization":   []string{`MediaBrowser Token="full-token"`},
		"Content-Length":  []string{"5171"},
		"Accept-Encoding": []string{"gzip"},
	})

	if got.Get("Authorization") != `MediaBrowser Token="full-token"` {
		t.Fatalf("authorization header was not preserved: %q", got.Get("Authorization"))
	}
	if got.Get("X-Emby-Token") != "" {
		t.Fatalf("unexpected fallback token header: %q", got.Get("X-Emby-Token"))
	}
	if got.Get("Content-Length") != "" {
		t.Fatalf("stale content length was preserved: %q", got.Get("Content-Length"))
	}
	if got.Get("Accept-Encoding") != "" {
		t.Fatalf("accept encoding was preserved: %q", got.Get("Accept-Encoding"))
	}
	if got.Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content type: %q", got.Get("Content-Type"))
	}
}

func TestBuildPlaybackInfoRequestHeaderAddsTokenFallback(t *testing.T) {
	got := buildPlaybackInfoRequestHeader(
		ItemInfo{ApiKeyType: Query, ApiKey: "query-token"},
		nil,
	)

	if got.Get(QueryTokenName) != "query-token" {
		t.Fatalf("expected fallback token header, got %q", got.Get(QueryTokenName))
	}
}

func TestRawFetchReturnsOriginStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("null"))
	}))
	defer server.Close()

	previousConfig := config.C
	config.C = &config.Config{Emby: &config.Emby{Host: server.URL}}
	defer func() { config.C = previousConfig }()

	res, _ := RawFetch("/Items/test/PlaybackInfo", http.MethodGet, nil, nil)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d with message %q", http.StatusUnauthorized, res.Code, res.Msg)
	}
}
