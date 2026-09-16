package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type claudeStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type apiWrapper struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type commandOptions struct {
	model       string
	prompt      string
	system      string
	output      string
	stream      bool
	maxTokens   int
	temperature float64
	topP        float64
}

type boolFlag struct {
	value bool
	set   bool
}

func (f *boolFlag) String() string {
	return fmt.Sprintf("%t", f.value)
}

func (f *boolFlag) Set(value string) error {
	switch strings.ToLower(value) {
	case "1", "t", "true", "y", "yes", "on":
		f.value = true
	case "0", "f", "false", "n", "no", "off":
		f.value = false
	default:
		return fmt.Errorf("invalid boolean value %q", value)
	}
	f.set = true
	return nil
}

func (f *boolFlag) IsBoolFlag() bool {
	return true
}

func promptFromArgs(args []string) string {
	prompt := strings.Join(args, " ")
	if prompt != "" {
		return prompt
	}
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stdin error: %v\n", err)
			os.Exit(1)
		}
		return strings.TrimSpace(string(b))
	}
	return ""
}

func outputWriter(path string) (io.Writer, func()) {
	if path == "" {
		return os.Stdout, func() {}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error: %v\n", err)
		os.Exit(1)
	}
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		fmt.Fprintf(os.Stderr, "file error: output path is a directory: %s\n", abs)
		os.Exit(1)
	}
	if dir := filepath.Dir(abs); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "file error: %v\n", err)
			os.Exit(1)
		}
	}
	f, err := os.Create(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error: %v\n", err)
		os.Exit(1)
	}
	return io.MultiWriter(os.Stdout, f), func() {
		f.Close()
		fmt.Fprintf(os.Stderr, "\nSaved to: \033]8;;file://%s\033\\%s\033]8;;\033\\\n", abs, abs)
	}
}

func printText(w io.Writer, text string) {
	if text == "" {
		return
	}
	fmt.Fprint(w, text)
	if !strings.HasSuffix(text, "\n") {
		fmt.Fprintln(w)
	}
}

func postJSON(endpoint string, body map[string]any, key string) *http.Response {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", apiBase+endpoint, bytes.NewReader(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "request error: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	if body["stream"] == true || strings.Contains(endpoint, ":streamGenerateContent") {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request error: %v\n", err)
		os.Exit(1)
	}
	return resp
}

func readJSONResponse(resp *http.Response) json.RawMessage {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "HTTP %d: %s\n", resp.StatusCode, string(raw))
		os.Exit(1)
	}

	var wrapped apiWrapper
	if json.Unmarshal(raw, &wrapped) == nil && wrapped.Data != nil {
		if wrapped.Code != 0 {
			fmt.Fprintf(os.Stderr, "API error %d: %s\n", wrapped.Code, wrapped.Message)
			os.Exit(1)
		}
		return wrapped.Data
	}

	var ae apiError
	if json.Unmarshal(raw, &ae) == nil && ae.Message != "" && ae.Code != 0 {
		if ae.Detail != "" {
			fmt.Fprintf(os.Stderr, "API error %d: %s: %s\n", ae.Code, ae.Message, ae.Detail)
		} else {
			fmt.Fprintf(os.Stderr, "API error %d: %s\n", ae.Code, ae.Message)
		}
		os.Exit(1)
	}
	return raw
}

func streamLines(resp *http.Response, onData func([]byte)) {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "HTTP %d: %s\n", resp.StatusCode, string(raw))
		os.Exit(1)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		if line == "[DONE]" {
			break
		}
		onData([]byte(line))
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stream error: %v\n", err)
		os.Exit(1)
	}
}

func commonFlags(command string, args []string, defaultModel string, defaultStream bool) commandOptions {
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	model := fs.String("model", defaultModel, "Model identifier")
	system := fs.String("system", "", "System instruction")
	output := fs.String("output", "", "Save generated text to this file path")
	fs.StringVar(output, "o", "", "Save generated text to this file path")
	stream := boolFlag{value: defaultStream}
	fs.Var(&stream, "stream", "Stream the response")
	maxTokens := fs.Int("max-tokens", 4096, "Maximum output tokens")
	temperature := fs.Float64("temperature", -1, "Sampling temperature; negative uses provider default")
	topP := fs.Float64("top-p", -1, "Nucleus sampling value; negative uses provider default")
	fs.Usage = func() {
		fmt.Printf("Usage: defapi text %s [flags] <prompt>\n", command)
		fmt.Println()
		fmt.Println("Prompt may also be piped on stdin.")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}

	return commandOptions{
		model:       *model,
		prompt:      prompt,
		system:      *system,
		output:      *output,
		stream:      stream.value,
		maxTokens:   *maxTokens,
		temperature: *temperature,
		topP:        *topP,
	}
}

func cmdOpenAI(command string, args []string) {
	opts := commonFlags(command, args, "openai/gpt-5.6-sol", false)
	key := apiKey()
	w, done := outputWriter(opts.output)
	defer done()

	messages := []chatMessage{}
	if opts.system != "" {
		messages = append(messages, chatMessage{Role: "system", Content: opts.system})
	}
	messages = append(messages, chatMessage{Role: "user", Content: opts.prompt})

	body := map[string]any{
		"model":    opts.model,
		"messages": messages,
		"stream":   opts.stream,
	}
	if opts.maxTokens > 0 {
		body["max_tokens"] = opts.maxTokens
	}
	if opts.temperature >= 0 {
		body["temperature"] = opts.temperature
	}
	if opts.topP >= 0 {
		body["top_p"] = opts.topP
	}

	resp := postJSON("/api/v1/chat/completions", body, key)
	if opts.stream {
		streamLines(resp, func(raw []byte) {
			var chunk openAIStreamChunk
			if json.Unmarshal(raw, &chunk) == nil {
				for _, choice := range chunk.Choices {
					fmt.Fprint(w, choice.Delta.Content)
				}
			}
		})
		fmt.Fprintln(w)
		return
	}

	var parsed openAIResponse
	if err := json.Unmarshal(readJSONResponse(resp), &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	if len(parsed.Choices) == 0 {
		fmt.Fprintln(os.Stderr, "error: no choices in response")
		os.Exit(1)
	}
	printText(w, parsed.Choices[0].Message.Content)
}

func cmdClaude(command string, args []string) {
	opts := commonFlags(command, args, "anthropic/claude-opus-5", false)
	key := apiKey()
	w, done := outputWriter(opts.output)
	defer done()

	body := map[string]any{
		"model":      opts.model,
		"messages":   []chatMessage{{Role: "user", Content: opts.prompt}},
		"max_tokens": opts.maxTokens,
		"stream":     opts.stream,
	}
	if opts.system != "" {
		body["system"] = opts.system
	}
	if opts.temperature >= 0 {
		body["temperature"] = opts.temperature
	}
	if opts.topP >= 0 {
		body["top_p"] = opts.topP
	}

	resp := postJSON("/api/v1/messages", body, key)
	if opts.stream {
		streamLines(resp, func(raw []byte) {
			var event claudeStreamEvent
			if json.Unmarshal(raw, &event) == nil && event.Type == "content_block_delta" {
				fmt.Fprint(w, event.Delta.Text)
			}
		})
		fmt.Fprintln(w)
		return
	}

	var parsed claudeResponse
	if err := json.Unmarshal(readJSONResponse(resp), &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	if len(parsed.Content) == 0 {
		fmt.Fprintln(os.Stderr, "error: no content in response")
		os.Exit(1)
	}
	for _, block := range parsed.Content {
		if block.Text != "" {
			printText(w, block.Text)
		}
	}
}

func geminiBody(opts commandOptions) map[string]any {
	config := map[string]any{"responseModalities": []string{"TEXT"}}
	if opts.maxTokens > 0 {
		config["maxOutputTokens"] = opts.maxTokens
	}
	if opts.temperature >= 0 {
		config["temperature"] = opts.temperature
	}
	if opts.topP >= 0 {
		config["topP"] = opts.topP
	}
	body := map[string]any{
		"contents": []map[string]any{{
			"role":  "user",
			"parts": []map[string]any{{"text": opts.prompt}},
		}},
		"generationConfig": config,
	}
	if opts.system != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]any{{"text": opts.system}},
		}
	}
	return body
}

func geminiEndpoint(model string, stream bool) string {
	endpoint := "/api/v1beta/models/" + url.PathEscape(strings.TrimPrefix(model, "google/"))
	if stream {
		return endpoint + ":streamGenerateContent?alt=sse"
	}
	return endpoint + ":generateContent"
}

func cmdGemini(command string, args []string) {
	opts := commonFlags(command, args, "gemini-3.8-flash", false)
	key := apiKey()
	w, done := outputWriter(opts.output)
	defer done()

	body := geminiBody(opts)
	endpoint := geminiEndpoint(opts.model, opts.stream)
	resp := postJSON(endpoint, body, key)
	if opts.stream {
		streamLines(resp, func(raw []byte) {
			var chunk geminiResponse
			if json.Unmarshal(raw, &chunk) == nil {
				for _, candidate := range chunk.Candidates {
					for _, part := range candidate.Content.Parts {
						fmt.Fprint(w, part.Text)
					}
				}
			}
		})
		fmt.Fprintln(w)
		return
	}

	var parsed geminiResponse
	if err := json.Unmarshal(readJSONResponse(resp), &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	if len(parsed.Candidates) == 0 {
		fmt.Fprintln(os.Stderr, "error: no candidates in response")
		os.Exit(1)
	}
	for _, part := range parsed.Candidates[0].Content.Parts {
		fmt.Fprint(w, part.Text)
	}
	fmt.Fprintln(w)
}

func textUsage() {
	fmt.Println(`Usage: defapi text <model> [flags] <prompt>

Models:
  gpt       OpenAI GPT-5.6 Sol via chat completions
  gemini    Google Gemini 3.8 Flash via generateContent
  claude    Anthropic Claude Opus 5 via messages

Aliases:
  openai, google, anthropic

Run 'defapi text <model> --help' for model-specific flags.`)
}

func cmdText(args []string) {
	if len(args) < 1 {
		textUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "gpt", "openai":
		cmdOpenAI(args[0], args[1:])
	case "gemini", "google":
		cmdGemini(args[0], args[1:])
	case "claude", "anthropic":
		cmdClaude(args[0], args[1:])
	case "-h", "--help", "help":
		textUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown model: %s\n\n", args[0])
		textUsage()
		os.Exit(1)
	}
}
