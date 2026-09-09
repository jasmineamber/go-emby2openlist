package emby

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/gin-gonic/gin"
)

func TestProxyLocalMediaKeepsOriginalRequestURI(t *testing.T) {
	var upstreamURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamURI = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	previousConfig := config.C
	config.C = &config.Config{Emby: &config.Emby{Host: server.URL}}
	defer func() { config.C = previousConfig }()

	req := httptest.NewRequest(http.MethodGet, "/Audio/item/universal?MediaSourceId=source&api_key=test", nil)
	req.Host = "public.example"
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req

	proxyLocalMedia(c, "/data/music/song.flac")

	if resp.Code != http.StatusOK {
		t.Fatalf("expected upstream status %d, got %d", http.StatusOK, resp.Code)
	}
	if upstreamURI != "/Audio/item/universal?MediaSourceId=source&api_key=test" {
		t.Fatalf("unexpected upstream URI: %q", upstreamURI)
	}
}

func TestProxyLocalMediaRedirectsOriginalURIForSameHost(t *testing.T) {
	previousConfig := config.C
	config.C = &config.Config{Emby: &config.Emby{Host: "http://192.168.2.21:8096"}}
	defer func() { config.C = previousConfig }()

	req := httptest.NewRequest(http.MethodGet, "/Videos/item/stream?Static=true", nil)
	req.Host = "192.168.2.21:8095"
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req

	proxyLocalMedia(c, "/data/movies/movie.mkv")

	if resp.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}
	if got := resp.Header().Get("Location"); got != "http://192.168.2.21:8096/Videos/item/stream?Static=true" {
		t.Fatalf("unexpected redirect location: %q", got)
	}
}
