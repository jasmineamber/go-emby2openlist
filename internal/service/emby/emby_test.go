package emby

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/gin-gonic/gin"
)

func TestProxySocketPreservesJellyfinSocketPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/socket" {
			t.Errorf("unexpected upstream path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "api_key=test" {
			t.Errorf("unexpected upstream query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer upstream.Close()

	previousConfig := config.C
	config.C = &config.Config{Emby: &config.Emby{Host: upstream.URL}}
	defer func() { config.C = previousConfig }()

	router := gin.New()
	router.Any("/*path", ProxySocket())
	proxyServer := httptest.NewServer(router)
	defer proxyServer.Close()

	resp, err := http.Get(proxyServer.URL + "/socket?api_key=test")
	if err != nil {
		t.Fatalf("request through socket proxy failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected upstream status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
