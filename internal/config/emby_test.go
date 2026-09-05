package config

import "testing"

func TestPlaybackConfigExcludesUserAgent(t *testing.T) {
	playback := &PlaybackConfig{
		ExcludeUserAgents: []string{" Hills Windows ", "Jellyfin Web", ""},
	}
	if err := playback.Init(); err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}

	tests := []struct {
		name      string
		userAgent string
		want      bool
	}{
		{name: "contains case insensitive", userAgent: "HILLS WINDOWS/1.4.1", want: true},
		{name: "second fragment", userAgent: "Jellyfin Web Chrome", want: true},
		{name: "not matched", userAgent: "Mozilla/5.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := playback.ExcludesUserAgent(tt.userAgent); got != tt.want {
				t.Fatalf("ExcludesUserAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmbyInitInitializesPlaybackConfig(t *testing.T) {
	emby := &Emby{Host: "http://localhost:8096"}
	if err := emby.Init(); err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}
	if emby.Playback == nil {
		t.Fatal("expected Playback config to be initialized")
	}
}
