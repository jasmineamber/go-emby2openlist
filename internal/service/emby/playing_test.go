package emby

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/jsons"
)

func TestSendPlayingProgressPreservesAuthorization(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get(HeaderAuthName) != `MediaBrowser Token="full-token"` {
			t.Errorf("authorization header was not preserved: %q", r.Header.Get(HeaderAuthName))
		}
		if r.ContentLength == 5171 {
			t.Errorf("stale content length was forwarded: %d", r.ContentLength)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	previousConfig := config.C
	config.C = &config.Config{Emby: &config.Emby{Host: server.URL}}
	defer func() { config.C = previousConfig }()

	body := jsons.NewEmptyObj()
	sendPlayingProgress(
		Query,
		QueryTokenName,
		"query-token",
		http.Header{
			HeaderAuthName:   []string{`MediaBrowser Token="full-token"`},
			"Content-Length": []string{"5171"},
		},
		body,
	)

	if requests != 2 {
		t.Fatalf("expected two progress requests, got %d", requests)
	}
}
