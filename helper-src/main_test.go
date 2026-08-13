package main

import "testing"

func TestValidHTTPURL(t *testing.T) {
	if !validHTTPURL("https://www.youtube.com/watch?v=abc") {
		t.Fatal("https URL rejected")
	}
	if validHTTPURL("file:///etc/passwd") {
		t.Fatal("file URL accepted")
	}
}

func TestBuildOptions(t *testing.T) {
	info := Info{Formats: []Format{
		{FormatID: "1", Ext: "mp4", VCodec: "avc1", ACodec: "none", Height: 1080, FPS: 60},
		{FormatID: "2", Ext: "m4a", VCodec: "none", ACodec: "mp4a"},
		{FormatID: "3", Ext: "mp4", VCodec: "avc1", ACodec: "mp4a", Height: 720, FPS: 30},
	}}
	opts := buildOptions(info)
	if len(opts) < 4 {
		t.Fatalf("expected several options, got %d", len(opts))
	}
	if opts[0].ID != "best" {
		t.Fatalf("first option should be best: %#v", opts[0])
	}
	found1080 := false
	foundAudio := false
	for _, o := range opts {
		if o.ID == "h:1080" {
			found1080 = true
		}
		if o.ID == "audio" {
			foundAudio = true
		}
	}
	if !found1080 || !foundAudio {
		t.Fatalf("missing quality/audio options: %#v", opts)
	}
}

func TestSelectors(t *testing.T) {
	for _, id := range []string{"best", "mp4", "audio", "h:1080", "h:720"} {
		if _, _, err := selector(id); err != nil {
			t.Fatalf("selector %s failed: %v", id, err)
		}
	}
	if _, _, err := selector("h:abc"); err == nil {
		t.Fatal("bad selector accepted")
	}
}

func TestStripRangeParams(t *testing.T) {
	got := stripRangeParams("https://cdn.example/video.mp4?foo=1&bytestart=0&byteend=999")
	if got != "https://cdn.example/video.mp4?foo=1" {
		t.Fatalf("range params not stripped: %s", got)
	}
}
