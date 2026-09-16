package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func extractOutputArg(args []string) ([]string, string) {
	var cleaned []string
	var output string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			cleaned = append(cleaned, args[i:]...)
			return cleaned, output
		case arg == "--output" || arg == "-o":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "error: %s requires a file path\n", arg)
				os.Exit(1)
			}
			output = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output="):
			output = strings.TrimPrefix(arg, "--output=")
		case strings.HasPrefix(arg, "-o="):
			output = strings.TrimPrefix(arg, "-o=")
		default:
			cleaned = append(cleaned, arg)
		}
	}
	return cleaned, output
}

func normalizeGPTImageModel(model string) string {
	return strings.TrimPrefix(model, "openai/")
}

func contains(items []string, item string) bool {
	for _, candidate := range items {
		if candidate == item {
			return true
		}
	}
	return false
}

func outputFlag(fs *flag.FlagSet) *string {
	output := fs.String("output", "", "Save image to this file path")
	fs.StringVar(output, "o", "", "Save image to this file path")
	return output
}

func cmdWan(args []string) {
	fs := flag.NewFlagSet("wan", flag.ExitOnError)
	ratio := fs.String("ratio", "1:1", "Aspect ratio: 1:1 16:9 4:3 21:9 3:4 9:16 8:1")
	output := outputFlag(fs)
	open := fs.Bool("open", false, "Open the image after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi image wan [flags] <prompt>")
		fs.PrintDefaults()
	}
	args, outputArg := extractOutputArg(args)
	fs.Parse(args)
	if outputArg != "" {
		*output = outputArg
	}

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}

	key := apiKey()
	fmt.Printf("Model: wan-2.7-image | Ratio: %s\nPrompt: %s\n\n", *ratio, prompt)

	data := post("/api/wan-image/gen", map[string]any{
		"model":        "wan-2.7-image",
		"prompt":       prompt,
		"aspect_ratio": *ratio,
	}, key)

	taskID := extractTaskID(data)
	imageURL := pollImage(taskID, key)
	dest := download(imageURL, taskID, "", *output)
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

func cmdMidjourney(args []string) {
	fs := flag.NewFlagSet("mj", flag.ExitOnError)
	speed := fs.String("speed", "fast", "Processing speed: fast, relax")
	bot := fs.String("bot", "MID_JOURNEY", "Bot type: MID_JOURNEY, NIJI_JOURNEY")
	image := fs.String("image", "", "Image URL or base64 for editing (uses edits endpoint)")
	output := outputFlag(fs)
	open := fs.Bool("open", false, "Open the image after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi image mj [flags] <prompt>")
		fmt.Println()
		fmt.Println("Midjourney parameters (--ar, --stylize, etc.) can be appended directly to the prompt.")
		fs.PrintDefaults()
	}
	args, outputArg := extractOutputArg(args)
	fs.Parse(args)
	if outputArg != "" {
		*output = outputArg
	}

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}

	key := apiKey()

	if *image != "" {
		fmt.Printf("Model: midjourney/edits | Speed: %s\nPrompt: %s\nImage: %s\n\n", *speed, prompt, *image)
		data := post("/api/midjourney/edits", map[string]any{
			"prompt": prompt,
			"image":  *image,
			"speed":  *speed,
		}, key)
		taskID := extractTaskID(data)
		imageURL := pollImage(taskID, key)
		dest := download(imageURL, taskID, "", *output)
		printResult(dest)
		if *open {
			openFile(dest)
		}
	} else {
		fmt.Printf("Model: midjourney/imagine | Bot: %s | Speed: %s\nPrompt: %s\n\n", *bot, *speed, prompt)
		data := post("/api/midjourney/imagine", map[string]any{
			"prompt":   prompt,
			"bot_type": *bot,
			"speed":    *speed,
		}, key)
		taskID := extractTaskID(data)
		imageURL := pollImage(taskID, key)
		dest := download(imageURL, taskID, "", *output)
		printResult(dest)
		if *open {
			openFile(dest)
		}
	}
}

func cmdGPTImage(command string, args []string, defaultModel string) {
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	model := fs.String("model", defaultModel, "Model: gpt-image-1.5, gpt-image-2")
	size := fs.String("size", "auto", "Output size: auto, 1024x1024, 1536x1024, 1024x1536 (gpt-image-1.5) or auto, 1:1, 16:9, 9:16 (gpt-image-2)")
	quality := fs.String("quality", "auto", "Quality: auto, high, medium, low")
	background := fs.String("background", "auto", "Background: auto, opaque, transparent")
	format := fs.String("format", "png", "Output format: png, jpeg, webp")
	var images stringListFlag
	fs.Var(&images, "image", "Reference image URL for gpt-image-2 editing (repeatable)")
	output := outputFlag(fs)
	open := fs.Bool("open", false, "Open the image after download")
	fs.Usage = func() {
		fmt.Printf("Usage: defapi image %s [flags] <prompt>\n", command)
		fs.PrintDefaults()
	}
	args, outputArg := extractOutputArg(args)
	fs.Parse(args)
	if outputArg != "" {
		*output = outputArg
	}

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}

	*model = normalizeGPTImageModel(*model)
	body := map[string]any{
		"model":  "openai/" + *model,
		"prompt": prompt,
		"size":   *size,
	}

	switch *model {
	case "gpt-image-1.5":
		if len(images) > 0 {
			fmt.Fprintln(os.Stderr, "error: --image is only supported with gpt-image-2")
			os.Exit(1)
		}
		body["quality"] = *quality
		body["background"] = *background
		body["output_format"] = *format
	case "gpt-image-2":
		if !contains([]string{"auto", "1:1", "16:9", "9:16"}, *size) {
			fmt.Fprintln(os.Stderr, "error: gpt-image-2 --size must be one of: auto, 1:1, 16:9, 9:16")
			os.Exit(1)
		}
		if len(images) > 16 {
			fmt.Fprintln(os.Stderr, "error: gpt-image-2 supports at most 16 --image values")
			os.Exit(1)
		}
		if len(images) > 0 {
			body["images"] = []string(images)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown GPT image model: %s\n", *model)
		os.Exit(1)
	}

	fmt.Printf("Model: %s | Size: %s", *model, *size)
	if *model == "gpt-image-1.5" {
		fmt.Printf(" | Quality: %s | Format: %s", *quality, *format)
	}
	if len(images) > 0 {
		fmt.Printf(" | Images: %d", len(images))
	}
	fmt.Printf("\nPrompt: %s\n\n", prompt)

	key := apiKey()
	data := post("/api/gpt-image/gen", body, key)

	taskID := extractTaskID(data)
	imageURL := pollImage(taskID, key)
	dest := download(imageURL, taskID, "", *output)
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

var googleModels = []string{
	"nano-banana",
	"nano-banana-pro",
	"nano-banana-2",
	"gemini-2.5-flash-image",
	"gemini-3.1-flash-image-preview",
}

var sizeSupportedModels = map[string]bool{
	"nano-banana-pro": true,
	"nano-banana-2":   true,
}

func cmdGoogle(args []string) {
	fs := flag.NewFlagSet("google", flag.ExitOnError)
	model := fs.String("model", "nano-banana-2", "Model: "+strings.Join(googleModels, ", "))
	ratio := fs.String("ratio", "1:1", "Aspect ratio: auto 1:1 16:9 21:9 2:3 3:2 3:4 4:3 4:5 5:4 9:16")
	size := fs.String("size", "", "Output resolution: 1k, 2k, 4k (only for nano-banana-pro and nano-banana-2)")
	output := outputFlag(fs)
	open := fs.Bool("open", false, "Open the image after download")
	fs.Usage = func() {
		fmt.Println("Usage: defapi image google [flags] <prompt>")
		fs.PrintDefaults()
	}
	args, outputArg := extractOutputArg(args)
	fs.Parse(args)
	if outputArg != "" {
		*output = outputArg
	}

	prompt := promptFromArgs(fs.Args())
	if prompt == "" {
		fs.Usage()
		os.Exit(1)
	}
	if *size != "" && !sizeSupportedModels[*model] {
		fmt.Fprintf(os.Stderr, "warning: --size is only supported by nano-banana-pro and nano-banana-2, ignoring\n")
		*size = ""
	}

	key := apiKey()
	fmt.Printf("Model: google/%s | Ratio: %s", *model, *ratio)
	if *size != "" {
		fmt.Printf(" | Size: %s", *size)
	}
	fmt.Printf("\nPrompt: %s\n\n", prompt)

	body := map[string]any{
		"model":        "google/" + *model,
		"prompt":       prompt,
		"aspect_ratio": *ratio,
	}
	if *size != "" {
		body["image_size"] = *size
	}

	data := post("/api/image/gen", body, key)
	taskID := extractTaskID(data)
	imageURL := pollImage(taskID, key)
	dest := download(imageURL, taskID, "", *output)
	printResult(dest)
	if *open {
		openFile(dest)
	}
}

func imageUsage() {
	fmt.Println(`Usage: defapi image <model> [flags] <prompt>

Models:
  wan     Alibaba Wan 2.7 Image (text-to-image)
  mj      Midjourney Imagine (text-to-image, or edit with --image)
  gpt     OpenAI GPT-Image-1.5/2 (text-to-image, or gpt-image-2 edit with --image)
  gpt2    OpenAI GPT-Image-2 shortcut
  google  Google image models via --model flag (default: nano-banana-2)
            nano-banana, nano-banana-pro, nano-banana-2,
            gemini-2.5-flash-image, gemini-3.1-flash-image-preview

Run 'defapi image <model> --help' for model-specific flags.`)
}

func cmdImage(args []string) {
	if len(args) < 1 {
		imageUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "wan":
		cmdWan(args[1:])
	case "mj", "midjourney":
		cmdMidjourney(args[1:])
	case "gpt":
		cmdGPTImage("gpt", args[1:], "gpt-image-1.5")
	case "gpt2":
		cmdGPTImage("gpt2", args[1:], "gpt-image-2")
	case "google":
		cmdGoogle(args[1:])
	case "-h", "--help", "help":
		imageUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown model: %s\n\n", args[0])
		imageUsage()
		os.Exit(1)
	}
}
