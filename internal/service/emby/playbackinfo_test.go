package emby

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/gin-gonic/gin"
)

func TestTransferPlaybackInfoPathRewriteByUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const sourcePath = "https://origin.example/static/%2Ffoo%2Fmovie.mkv"
	var playbackRequests atomic.Int32
	playbackDone := make(chan struct{})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("IsPlayback") == "true" && playbackRequests.Add(1) == 2 {
			close(playbackDone)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"MediaSources":[{"Id":"source-id","Path":%q,"IsRemote":true}],"PlaySessionId":"session"}`, sourcePath)
	}))
	defer origin.Close()

	previousConfig := config.C
	config.C = &config.Config{
		Emby: &config.Emby{
			Host: origin.URL,
			Strm: &config.Strm{
				PathMap: []string{
					sourcePath[:strings.Index(sourcePath, "movie.mkv")] + " => https://mapped.example/d/foo/",
				},
			},
			Playback: &config.PlaybackConfig{
				ExcludeUserAgents: []string{"Hills Windows"},
			},
		},
		Openlist: &config.Openlist{EnableSign: false},
		Cache:    &config.Cache{Enable: false},
	}
	if err := config.C.Emby.Strm.Init(); err != nil {
		t.Fatalf("unexpected strm init error: %v", err)
	}
	if err := config.C.Emby.Playback.Init(); err != nil {
		t.Fatalf("unexpected playback init error: %v", err)
	}
	defer func() { config.C = previousConfig }()

	requestPath := func(userAgent string) string {
		req := httptest.NewRequest(http.MethodPost, "/Items/item-id/PlaybackInfo?api_key=test-token", strings.NewReader(`{}`))
		req.Header.Set("User-Agent", userAgent)
		resp := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(resp)
		ctx.Request = req

		TransferPlaybackInfo(ctx)
		if resp.Code != http.StatusOK {
			t.Fatalf("unexpected response status: %d, body: %s", resp.Code, resp.Body.String())
		}

		var body struct {
			MediaSources []struct {
				Path string
			} `json:"MediaSources"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(body.MediaSources) != 1 {
			t.Fatalf("unexpected media source count: %d", len(body.MediaSources))
		}
		return body.MediaSources[0].Path
	}

	if got := requestPath("Mozilla/5.0"); got != "https://mapped.example/d/foo/movie.mkv" {
		t.Fatalf("mapped Path = %q", got)
	}
	if got := requestPath("HILLS WINDOWS/1.4.1"); got != "https://origin.example/static//foo/movie.mkv" {
		t.Fatalf("excluded Path = %q", got)
	}

	select {
	case <-playbackDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for remote playback notifications")
	}
}

func TestCalcPlaybackInfoSpaceCacheKeySeparatesPathModes(t *testing.T) {
	itemInfo := ItemInfo{Id: "item-id", ApiKey: "api-key"}
	mapped := calcPlaybackInfoSpaceCacheKey(itemInfo, false)
	original := calcPlaybackInfoSpaceCacheKey(itemInfo, true)
	if mapped == original {
		t.Fatalf("path modes share cache key: %q", mapped)
	}
	if !strings.HasSuffix(mapped, "_path-mapped") || !strings.HasSuffix(original, "_path-original") {
		t.Fatalf("unexpected cache keys: %q, %q", mapped, original)
	}
}
