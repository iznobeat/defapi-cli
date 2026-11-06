package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestKlingBodyReferenceMode(t *testing.T) {
	body, err := klingBody(klingOptions{
		prompt: "@image_1 smiles", duration: 10, resolution: "1080p", ratio: "9:16",
		images: []string{"https://x/a.png"}, videos: []string{"https://x/v.mp4"}, keepSound: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"model": "kling-4.0", "mode": "reference-to-video", "prompt": "@image_1 smiles",
		"duration": 10, "resolution": "1080p", "aspect_ratio": "9:16",
		"image_urls": []string{"https://x/a.png"}, "video_urls": []string{"https://x/v.mp4"},
		"keep_original_sound": true,
	}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %#v", body)
	}
}

func TestKlingBodyDefaultsOmitOptionalFields(t *testing.T) {
	body, err := klingBody(klingOptions{prompt: "cat", duration: 5, resolution: "720p"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"model": "kling-4.0", "mode": "reference-to-video", "prompt": "cat", "duration": 5, "resolution": "720p"}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %#v", body)
	}
}

func TestKlingBodyKeyframeMode(t *testing.T) {
	body, err := klingBody(klingOptions{
		prompt: "push in", duration: 5, resolution: "720p",
		firstFrame: "https://x/1.png", keyframes: []string{"https://x/2.png"}, lastFrame: "https://x/3.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"model": "kling-4.0", "mode": "keyframe-to-video", "prompt": "push in", "duration": 5, "resolution": "720p",
		"first_frame_url": "https://x/1.png", "keyframe_urls": []string{"https://x/2.png"}, "last_frame_url": "https://x/3.png",
	}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %#v", body)
	}
}

func TestKlingBodyRejectsInvalidOptions(t *testing.T) {
	frames := func(n int) []string { return make([]string, n) }
	for _, tc := range []struct {
		name string
		opts klingOptions
		want string
	}{
		{"short", klingOptions{duration: 2, resolution: "720p"}, "--duration"},
		{"long", klingOptions{duration: 31, resolution: "720p"}, "--duration"},
		{"resolution", klingOptions{duration: 5, resolution: "4k"}, "--resolution"},
		{"ratio", klingOptions{duration: 5, resolution: "720p", ratio: "4:3"}, "--ratio"},
		{"mixed modes", klingOptions{duration: 5, resolution: "720p", firstFrame: "f", images: []string{"i"}}, "cannot be combined"},
		{"no first frame", klingOptions{duration: 5, resolution: "720p", lastFrame: "l"}, "--first-frame is required"},
		{"frame cap", klingOptions{duration: 3, resolution: "720p", firstFrame: "f", keyframes: frames(4)}, "at most 4 allowed"},
		{"too many images", klingOptions{duration: 5, resolution: "720p", images: frames(11)}, "--image"},
		{"too many videos", klingOptions{duration: 5, resolution: "720p", videos: frames(6)}, "--video"},
		{"keep sound alone", klingOptions{duration: 5, resolution: "720p", keepSound: true}, "--keep-sound requires"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := klingBody(tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}
