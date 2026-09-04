package emby

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
)

func TestGetOpenlistSignedLink(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/fs/get" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "openlist-token" {
			t.Errorf("unexpected authorization header: %s", got)
		}

		var body struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		requestedPath = body.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"message":"","data":{"raw_url":"https://cdn.example/file.mkv?sign=test"}}`))
	}))
	defer server.Close()

	previousConfig := config.C
	config.C = &config.Config{
		Openlist: &config.Openlist{
			Host:  server.URL,
			Token: "openlist-token",
		},
	}
	defer func() { config.C = previousConfig }()

	got := getOpenlistSignedLink(
		"https://openlist.example/d/115/%E8%A7%86%E5%90%AC%E5%A8%B1%E4%B9%90%2F%E7%94%B5%E5%BD%B1%2Ffile.mkv?source=test",
		http.Header{"User-Agent": []string{"test-agent"}},
	)

	if got != "https://cdn.example/file.mkv?sign=test" {
		t.Fatalf("unexpected signed link: %s", got)
	}
	if requestedPath != "/115/视听娱乐/电影/file.mkv" {
		t.Fatalf("unexpected OpenList path: %s", requestedPath)
	}
}

func TestGetOpenlistSignedLinkFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	previousConfig := config.C
	config.C = &config.Config{
		Openlist: &config.Openlist{
			Host:  server.URL,
			Token: "openlist-token",
		},
	}
	defer func() { config.C = previousConfig }()

	originLink := "https://openlist.example/115/file.mkv"
	if got := getOpenlistSignedLink(originLink, nil); got != originLink {
		t.Fatalf("expected fallback link %q, got %q", originLink, got)
	}
}

func TestGetOpenlistSignedLinkInvalidURLFallback(t *testing.T) {
	previousConfig := config.C
	config.C = &config.Config{Openlist: &config.Openlist{}}
	defer func() { config.C = previousConfig }()

	originLink := "://invalid-url"
	if got := getOpenlistSignedLink(originLink, nil); got != originLink {
		t.Fatalf("expected fallback link %q, got %q", originLink, got)
	}
}
