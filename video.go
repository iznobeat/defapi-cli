package main

import (
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

func videoUsage() {
	fmt.Println(`Usage: defapi video <model> [flags] <prompt>

Models:
  seedance   ByteDance Seedance 2.0
  grok       xAI Grok Imagine Video
  sora       OpenAI Sora 2 Stable

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
	case "-h", "--help", "help":
		videoUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown model: %s\n\n", args[0])
		videoUsage()
		os.Exit(1)
	}
}
