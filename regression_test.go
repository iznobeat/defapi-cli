package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func mockHTTP(t *testing.T, f roundTripFunc) {
	t.Helper()
	old := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: f}
	t.Cleanup(func() { http.DefaultClient = old })
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestDownloadPreservesOutputOnFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   io.Reader
	}{
		{"HTTP error", 403, strings.NewReader("denied")},
		{"truncated body", 200, io.MultiReader(strings.NewReader("partial"), brokenReader{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			dest := filepath.Join(dir, "image.png")
			if err := os.WriteFile(dest, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			mockHTTP(t, func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(tc.body)}, nil
			})
			if err := downloadFile("https://media.example/image.png", dest); err == nil {
				t.Fatal("expected download failure")
			}
			got, err := os.ReadFile(dest)
			if err != nil || string(got) != "original" {
				t.Fatalf("existing output changed: %q, %v", got, err)
			}
			files, err := os.ReadDir(dir)
			if err != nil || len(files) != 1 {
				t.Fatalf("temporary file left behind: %v, %v", files, err)
			}
		})
	}
}

func TestDownloadSuccess(t *testing.T) {
	mockHTTP(t, func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("complete image"))}, nil
	})
	dest := filepath.Join(t.TempDir(), "image.png")
	if err := downloadFile("https://media.example/image.png", dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "complete image" {
		t.Fatalf("output = %q, %v", got, err)
	}
}

func TestReadResponseDataRejectsTruncatedBody(t *testing.T) {
	resp := &http.Response{StatusCode: 200, Body: io.NopCloser(io.MultiReader(strings.NewReader(`{"code":0,"data":{}}`), brokenReader{}))}
	if _, err := readResponseData(resp); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}

func TestGuessExtUsesURLPath(t *testing.T) {
	for raw, want := range map[string]string{
		"https://cdn.example/image.JPG?fallback=.png": ".jpg",
		"https://cdn.example.png/image.webp":          ".webp",
		"https://cdn.example/image?name=file.jpg":     ".png",
	} {
		if got := guessExt(raw); got != want {
			t.Errorf("guessExt(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestGeminiRequestOptions(t *testing.T) {
	opts := commandOptions{prompt: "article", system: "summarize", maxTokens: 123, temperature: 0, topP: 0.9}
	body := geminiBody(opts)
	wantConfig := map[string]any{"responseModalities": []string{"TEXT"}, "maxOutputTokens": 123, "temperature": float64(0), "topP": 0.9}
	if !reflect.DeepEqual(body["generationConfig"], wantConfig) {
		t.Fatalf("config = %#v", body["generationConfig"])
	}
	wantContents := []map[string]any{{"role": "user", "parts": []map[string]any{{"text": "article"}}}}
	if !reflect.DeepEqual(body["contents"], wantContents) {
		t.Fatalf("contents = %#v", body["contents"])
	}
	wantSystem := map[string]any{"parts": []map[string]any{{"text": "summarize"}}}
	if !reflect.DeepEqual(body["systemInstruction"], wantSystem) {
		t.Fatalf("systemInstruction = %#v", body["systemInstruction"])
	}
	body = geminiBody(commandOptions{prompt: "hello", temperature: -1, topP: -1})
	if _, exists := body["systemInstruction"]; exists {
		t.Fatal("unexpected system instruction")
	}
	if len(body["generationConfig"].(map[string]any)) != 1 {
		t.Fatalf("unexpected defaults: %#v", body)
	}
}

func TestGeminiStreamingRequest(t *testing.T) {
	mockHTTP(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/api/v1beta/models/gemini-3.8-flash:streamGenerateContent" || req.URL.Query().Get("alt") != "sse" {
			t.Errorf("stream URL = %s", req.URL)
		}
		if got := req.Header.Get("Accept"); got != "text/event-stream" {
			t.Errorf("Accept = %q", got)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("data: {\"text\":\"hello\"}\n\ndata: [DONE]\n"))}, nil
	})
	resp := postJSON(geminiEndpoint("google/gemini-3.8-flash", true), geminiBody(commandOptions{prompt: "hi"}), "test-key")
	var chunks []string
	streamLines(resp, func(raw []byte) { chunks = append(chunks, string(raw)) })
	if !reflect.DeepEqual(chunks, []string{`{"text":"hello"}`}) {
		t.Fatalf("chunks = %#v", chunks)
	}
	if got := geminiEndpoint("google/gemini-3.8-flash", false); got != "/api/v1beta/models/gemini-3.8-flash:generateContent" {
		t.Fatalf("endpoint = %s", got)
	}
}

func TestOutputFlagRespectsTerminator(t *testing.T) {
	args := []string{"--", "--output", "literal prompt"}
	got, output := extractOutputArg(args)
	if output != "" || !reflect.DeepEqual(got, args) {
		t.Fatalf("got %q, output %q", got, output)
	}
}
