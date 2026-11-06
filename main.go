package main

import (
	"fmt"
	"os"
)

func mainUsage() {
	fmt.Println(`Usage: defapi <category> <model> [flags] <prompt>

Categories:
  text    Text generation (gpt, gemini, claude)
  image   Image generation (wan, mj, gpt, gpt2, google)
  video   Video generation (seedance, grok, sora, kling)

Run 'defapi <category> --help' for model-specific usage.`)
}

func main() {
	if len(os.Args) < 2 {
		mainUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "text":
		cmdText(os.Args[2:])
	case "image":
		cmdImage(os.Args[2:])
	case "video":
		cmdVideo(os.Args[2:])
	case "-h", "--help", "help":
		mainUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown category: %s\n\n", os.Args[1])
		mainUsage()
		os.Exit(1)
	}
}
