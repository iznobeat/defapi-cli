package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	apiBase                = "https://api.defapi.org"
	pollInterval           = 5 * time.Second
	taskNotFoundRetryLimit = 5
)

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type responseError struct {
	StatusCode int
	Body       string
	Code       int
	Message    string
}

func (e *responseError) Error() string {
	if e.StatusCode >= 400 {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

type taskData struct {
	TaskID       string          `json:"task_id"`
	Status       string          `json:"status"`
	Result       json.RawMessage `json:"result"`
	StatusReason struct {
		Message *string `json:"message"`
	} `json:"status_reason"`
}

func apiKey() string {
	key := os.Getenv("DEFAPI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "error: DEFAPI_API_KEY environment variable not set")
		os.Exit(1)
	}
	return key
}

func post(endpoint string, body map[string]any, key string) json.RawMessage {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", apiBase+endpoint, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	return readResponse(resp)
}

func getData(endpoint, key string) (json.RawMessage, error) {
	req, _ := http.NewRequest("GET", apiBase+endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()
	return readResponseData(resp)
}

func readResponse(resp *http.Response) json.RawMessage {
	data, err := readResponseData(resp)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return data
}

func readResponseData(resp *http.Response) (json.RawMessage, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	var ar apiResponse
	if err := json.Unmarshal(raw, &ar); err != nil && resp.StatusCode < 400 {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, &responseError{
			StatusCode: resp.StatusCode,
			Body:       string(raw),
			Code:       ar.Code,
			Message:    ar.Message,
		}
	}
	if ar.Code != 0 {
		return nil, &responseError{Code: ar.Code, Message: ar.Message}
	}
	return ar.Data, nil
}

func isTaskNotFoundError(err error) bool {
	var re *responseError
	if !errors.As(err, &re) {
		return false
	}
	return re.StatusCode == http.StatusNotFound && re.Code == 1 && re.Message == "task not found"
}

func extractTaskID(data json.RawMessage) string {
	var d struct {
		TaskID string `json:"task_id"`
	}
	json.Unmarshal(data, &d)
	return d.TaskID
}

func pollLoop(taskID, key string) json.RawMessage {
	if taskID == "" {
		fmt.Fprintln(os.Stderr, "error: no task ID in submission response")
		os.Exit(1)
	}
	fmt.Printf("Task submitted: %s\nPolling", taskID)
	taskNotFoundRetries := 0
	for {
		time.Sleep(pollInterval)
		data, err := getData("/api/task/query?task_id="+url.QueryEscape(taskID), key)
		if err != nil {
			if isTaskNotFoundError(err) && taskNotFoundRetries < taskNotFoundRetryLimit {
				taskNotFoundRetries++
				fmt.Print(".")
				continue
			}
			fmt.Fprintf(os.Stderr, "\n%v\n", err)
			os.Exit(1)
		}
		taskNotFoundRetries = 0

		var td taskData
		if err := json.Unmarshal(data, &td); err != nil || td.Status == "" {
			fmt.Fprintln(os.Stderr, "error: invalid task status response")
			os.Exit(1)
		}

		switch td.Status {
		case "success":
			fmt.Println(" done.")
			return td.Result
		case "failed":
			msg := "unknown reason"
			if td.StatusReason.Message != nil && *td.StatusReason.Message != "" {
				msg = *td.StatusReason.Message
			}
			pretty, _ := json.MarshalIndent(json.RawMessage(data), "", "  ")
			fmt.Fprintf(os.Stderr, "\ngeneration failed: %s\n%s\n", msg, pretty)
			os.Exit(1)
		default:
			fmt.Print(".")
		}
	}
}

func extractImageURL(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var arr []struct {
		Image string `json:"image"`
	}
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		return arr[0].Image
	}
	var obj struct {
		Image       string `json:"image"`
		BigImageURL string `json:"big_image_url"`
	}
	json.Unmarshal(raw, &obj)
	if obj.Image != "" {
		return obj.Image
	}
	return obj.BigImageURL
}

func pollImage(taskID, key string) string {
	result := pollLoop(taskID, key)
	imageURL := extractImageURL(result)
	if imageURL == "" {
		fmt.Fprintln(os.Stderr, "error: no image URL in response")
		os.Exit(1)
	}
	fmt.Printf("Image URL: %s\n", imageURL)
	return imageURL
}

func pollVideo(taskID, key string) string {
	result := pollLoop(taskID, key)
	var r struct {
		Video string `json:"video"`
	}
	json.Unmarshal(result, &r)
	if r.Video == "" {
		fmt.Fprintln(os.Stderr, "error: no video URL in response")
		os.Exit(1)
	}
	return r.Video
}

func guessExt(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ".png"
	}
	pathExt := strings.ToLower(filepath.Ext(u.Path))
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
		if pathExt == ext {
			return ext
		}
	}
	return ".png"
}

// download saves mediaURL to outputPath, or ~/Downloads/defapi_<taskID><ext> if empty.
// Pass defaultExt="" to guess from URL (images); pass ".mp4" for videos.
func download(mediaURL, taskID, defaultExt, outputPath string) string {
	dest := outputPath
	if dest == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "home directory error: %v\n", err)
			os.Exit(1)
		}
		dir := filepath.Join(home, "Downloads")
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "file error: %v\n", err)
			os.Exit(1)
		}
		ext := defaultExt
		if ext == "" {
			ext = guessExt(mediaURL)
		}
		dest = filepath.Join(dir, "defapi_"+taskID+ext)
	} else {
		abs, err := filepath.Abs(dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "file error: %v\n", err)
			os.Exit(1)
		}
		dest = abs
		if info, err := os.Stat(dest); err == nil && info.IsDir() {
			fmt.Fprintf(os.Stderr, "file error: output path is a directory: %s\n", dest)
			os.Exit(1)
		}
		if dir := filepath.Dir(dest); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "file error: %v\n", err)
				os.Exit(1)
			}
		}
	}

	fmt.Println("Downloading...")
	if err := downloadFile(mediaURL, dest); err != nil {
		fmt.Fprintf(os.Stderr, "download error: %v\n", err)
		os.Exit(1)
	}
	return dest
}

// Stage downloads beside the destination so failures preserve existing output.
func downloadFile(mediaURL, dest string) error {
	resp, err := http.Get(mediaURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.CreateTemp(filepath.Dir(dest), ".defapi-download-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), dest)
}

func printResult(dest string) {
	fmt.Printf("\nSaved to: \033]8;;file://%s\033\\%s\033]8;;\033\\\n", dest, dest)
}

func openFile(path string) {
	cmd := "xdg-open"
	if runtime.GOOS == "darwin" {
		cmd = "open"
	}
	if err := exec.Command(cmd, path).Start(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not open file: %v\n", err)
	}
}
