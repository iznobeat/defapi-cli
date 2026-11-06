package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func cmdSeedance(args []string) {
	fs := flag.NewFlagSet("seedance", flag.ExitOnError)
	duration := fs.Int("duration", 5, "Duration in seconds: 5, 10, 15")
	ratio := fs.String("ratio", "16:9", "Aspect ratio: 16:9 9:16 1:1 4:3 3:4 21:9")
	image := fs.String("image", "", "Reference image URL for image-to-video (up to 9 supported)")
	open := fs.Bool("open", false, "Open the video after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi video seedance [flags] <prompt>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}
	validDuration := map[int]bool{5: true, 10: true, 15: true}
	if !validDuration[*duration] {
		fmt.Fprintln(os.Stderr, "error: --duration must be 5, 10, or 15")
		os.Exit(1)
	}

	key := apiKey()
	fmt.Printf("Model: seedance | Duration: %ds | Ratio: %s\nPrompt: %s\n\n", *duration, *ratio, prompt)

	content := []map[string]any{{"type": "text", "text": prompt}}
	if *image != "" {
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]any{"url": *image},
		})
	}

	data := post("/api/video/seedance/gen", map[string]any{
		"model":    "seedance-2.0",
		"content":  content,
		"duration": *duration,
		"ratio":    *ratio,
	}, key)

	taskID := extractTaskID(data)
	videoURL := pollVideo(taskID, key)
	dest := download(videoURL, taskID, ".mp4", "")
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

func cmdGrok(args []string) {
	fs := flag.NewFlagSet("grok", flag.ExitOnError)
	duration := fs.Int("duration", 10, "Duration in seconds: 10, 15")
	ratio := fs.String("ratio", "16:9", "Aspect ratio: 16:9 9:16 1:1 2:3 3:2")
	image := fs.String("image", "", "Reference image URL for image-to-video")
	open := fs.Bool("open", false, "Open the video after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi video grok [flags] <prompt>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}
	validDuration := map[int]bool{10: true, 15: true}
	if !validDuration[*duration] {
		fmt.Fprintln(os.Stderr, "error: --duration must be 10 or 15")
		os.Exit(1)
	}

	key := apiKey()
	fmt.Printf("Model: grok | Duration: %ds | Ratio: %s\nPrompt: %s\n\n", *duration, *ratio, prompt)

	body := map[string]any{
		"prompt":       prompt,
		"model":        "grok-imagine-video",
		"duration":     fmt.Sprintf("%d", *duration),
		"aspect_ratio": *ratio,
	}
	if *image != "" {
		body["images"] = []string{*image}
	}

	data := post("/api/grok-imagine-video/gen", body, key)

	taskID := extractTaskID(data)
	videoURL := pollVideo(taskID, key)
	dest := download(videoURL, taskID, ".mp4", "")
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

func cmdSora(args []string) {
	fs := flag.NewFlagSet("sora", flag.ExitOnError)
	duration := fs.Int("duration", 10, "Duration in seconds: 10, 15, 25 (25 requires --variant sora-2-pro)")
	ratio := fs.String("ratio", "16:9", "Aspect ratio: 16:9 9:16")
	variant := fs.String("variant", "sora-2", "Model variant: sora-2, sora-2-hd, sora-2-pro")
	image := fs.String("image", "", "Reference image URL for image-to-video")
	open := fs.Bool("open", false, "Open the video after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi video sora [flags] <prompt>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}
	validDuration := map[int]bool{10: true, 15: true, 25: true}
	if !validDuration[*duration] {
		fmt.Fprintln(os.Stderr, "error: --duration must be 10, 15, or 25")
		os.Exit(1)
	}
	if *duration == 25 && *variant != "sora-2-pro" {
		fmt.Fprintln(os.Stderr, "warning: 25s duration requires sora-2-pro, switching variant")
		*variant = "sora-2-pro"
	}

	key := apiKey()
	fmt.Printf("Model: sora (%s) | Duration: %ds | Ratio: %s\nPrompt: %s\n\n", *variant, *duration, *ratio, prompt)

	body := map[string]any{
		"prompt":       prompt,
		"model":        *variant,
		"duration":     fmt.Sprintf("%d", *duration),
		"aspect_ratio": *ratio,
	}
	if *image != "" {
		body["images"] = []string{*image}
	}

	data := post("/api/sora2/gen", body, key)

	taskID := extractTaskID(data)
	videoURL := pollVideo(taskID, key)
	dest := download(videoURL, taskID, ".mp4", "")
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

type klingOptions struct {
	prompt     string
	duration   int
	resolution string
	ratio      string
	images     []string
	videos     []string
	keepSound  bool
	firstFrame string
	keyframes  []string
	lastFrame  string
}

// klingBody builds a /api/kling/gen request. Keyframe flags select
// keyframe-to-video; otherwise the default reference-to-video mode is used.
func klingBody(o klingOptions) (map[string]any, error) {
	if o.duration < 3 || o.duration > 30 {
		return nil, errors.New("--duration must be between 3 and 30")
	}
	if o.resolution != "720p" && o.resolution != "1080p" {
		return nil, errors.New("--resolution must be 720p or 1080p")
	}
	validRatio := map[string]bool{"": true, "auto": true, "16:9": true, "9:16": true, "1:1": true, "21:9": true}
	if !validRatio[o.ratio] {
		return nil, errors.New("--ratio must be auto, 16:9, 9:16, 1:1, or 21:9")
	}

	body := map[string]any{
		"model":      "kling-4.0",
		"prompt":     o.prompt,
		"duration":   o.duration,
		"resolution": o.resolution,
	}
	if o.ratio != "" {
		body["aspect_ratio"] = o.ratio
	}

	keyframeMode := o.firstFrame != "" || len(o.keyframes) > 0 || o.lastFrame != ""
	if keyframeMode {
		if len(o.images) > 0 || len(o.videos) > 0 || o.keepSound {
			return nil, errors.New("keyframe flags cannot be combined with --image, --video, or --keep-sound")
		}
		if o.firstFrame == "" {
			return nil, errors.New("--first-frame is required when using --keyframe or --last-frame")
		}
		if len(o.keyframes) > 8 {
			return nil, errors.New("at most 8 --keyframe values are allowed")
		}
		frames := 1 + len(o.keyframes)
		if o.lastFrame != "" {
			frames++
		}
		if limit := min(o.duration+1, 10); frames > limit {
			return nil, fmt.Errorf("%d frames given; at most %d allowed for a %ds video", frames, limit, o.duration)
		}
		body["mode"] = "keyframe-to-video"
		body["first_frame_url"] = o.firstFrame
		if len(o.keyframes) > 0 {
			body["keyframe_urls"] = o.keyframes
		}
		if o.lastFrame != "" {
			body["last_frame_url"] = o.lastFrame
		}
		return body, nil
	}

	if len(o.images) > 10 {
		return nil, errors.New("at most 10 --image values are allowed")
	}
	if len(o.videos) > 5 {
		return nil, errors.New("at most 5 --video values are allowed")
	}
	if o.keepSound && len(o.videos) == 0 {
		return nil, errors.New("--keep-sound requires --video")
	}
	body["mode"] = "reference-to-video"
	if len(o.images) > 0 {
		body["image_urls"] = o.images
	}
	if len(o.videos) > 0 {
		body["video_urls"] = o.videos
		if o.keepSound {
			body["keep_original_sound"] = true
		}
	}
	return body, nil
}

func cmdKling(args []string) {
	fs := flag.NewFlagSet("kling", flag.ExitOnError)
	var images, videos, keyframes stringListFlag
	duration := fs.Int("duration", 5, "Duration in seconds: 3-30")
	resolution := fs.String("resolution", "720p", "Resolution: 720p, 1080p")
	ratio := fs.String("ratio", "", "Aspect ratio: auto 16:9 9:16 1:1 21:9 (default: API decides)")
	fs.Var(&images, "image", "Reference image URL, referenced as @image_1... (repeatable, max 10)")
	fs.Var(&videos, "video", "Reference video URL, referenced as @video_1... (repeatable, max 5; billed for input length)")
	keepSound := fs.Bool("keep-sound", false, "Keep the reference video's original audio")
	firstFrame := fs.String("first-frame", "", "First frame image URL (enables keyframe-to-video)")
	fs.Var(&keyframes, "keyframe", "Intermediate keyframe image URL (repeatable, max 8)")
	lastFrame := fs.String("last-frame", "", "Last frame image URL")
	open := fs.Bool("open", false, "Open the video after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi video kling [flags] <prompt>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}
	body, err := klingBody(klingOptions{
		prompt:     prompt,
		duration:   *duration,
		resolution: *resolution,
		ratio:      *ratio,
		images:     images,
		videos:     videos,
		keepSound:  *keepSound,
		firstFrame: *firstFrame,
		keyframes:  keyframes,
		lastFrame:  *lastFrame,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	key := apiKey()
	fmt.Printf("Model: kling (%s) | Duration: %ds | Resolution: %s\nPrompt: %s\n\n", body["mode"], *duration, *resolution, prompt)

	data := post("/api/kling/gen", body, key)

	taskID := extractTaskID(data)
	videoURL := pollVideo(taskID, key)
	dest := download(videoURL, taskID, ".mp4", "")
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

func videoUsage() {
	fmt.Println(`Usage: defapi video <model> [flags] <prompt>

Models:
  seedance   ByteDance Seedance 2.0
  grok       xAI Grok Imagine Video
  sora       OpenAI Sora 2 Stable
  kling      Kuaishou Kling 4.0

Run 'defapi video <model> --help' for model-specific flags.`)
}

func cmdVideo(args []string) {
	if len(args) < 1 {
		videoUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "seedance":
		cmdSeedance(args[1:])
	case "grok":
		cmdGrok(args[1:])
	case "sora":
		cmdSora(args[1:])
	case "kling":
		cmdKling(args[1:])
	case "-h", "--help", "help":
		videoUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown model: %s\n\n", args[0])
		videoUsage()
		os.Exit(1)
	}
}
