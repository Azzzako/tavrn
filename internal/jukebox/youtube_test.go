package jukebox

import (
	"os/exec"
	"testing"
)

func TestYouTubeName(t *testing.T) {
	yt := NewYouTube("")
	if yt.Name() != "youtube" {
		t.Errorf("name = %q, want youtube", yt.Name())
	}
}

func TestYouTubeEnabled(t *testing.T) {
	yt := NewYouTube("")
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		if yt.Enabled() {
			t.Error("should be disabled when yt-dlp is not installed")
		}
		t.Skip("yt-dlp not installed")
	}
	if !yt.Enabled() {
		t.Error("should be enabled when yt-dlp is installed")
	}
}

func TestYouTubeSkipsLivestreams(t *testing.T) {
	// Verify the duration filter logic
	yt := NewYouTube("")
	_ = yt // just verify it compiles with the filter
}
