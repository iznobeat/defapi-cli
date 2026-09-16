# defapi

CLI for text, image, and video generation via [defapi.org](https://defapi.org).

**Requires a defapi.org account and API key** — sign up at [defapi.org](https://defapi.org), then create an API key in your account settings.

## Setup

Export your key:

```sh
export DEFAPI_API_KEY=your_api_key_here
```

Add to `~/.bashrc` or `~/.zshrc` to persist it.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/jhgundersen/defapi-cli/master/install.sh | sh
```

Or build from source with Go 1.26.2 or newer (the included `mise.toml` pins 1.26.2):

```sh
go build -o defapi .
```

## Usage

```
defapi <category> <model> [flags] <prompt>
```

Prompt can also be piped via stdin for all categories. A positional prompt takes
precedence over stdin; flags should precede the prompt.

### Text

```sh
defapi text gpt    "Explain recursion"
defapi text claude "Write a haiku about Go"
defapi text gemini --system "Summarize this article" < article.txt
```

| Model    | Provider  | Default                        |
|----------|-----------|--------------------------------|
| `gpt`    | OpenAI    | gpt-5.6-sol                    |
| `claude` | Anthropic | claude-opus-5                  |
| `gemini` | Google    | gemini-3.8-flash                |

Common flags: `--model`, `--system`, `--stream`, `--max-tokens`, `--temperature`, `--top-p`, `--output/-o`

These defaults can also be selected explicitly with `--model`:

```sh
defapi text gpt --model openai/gpt-5.6-sol "Explain recursion"
defapi text claude --model anthropic/claude-opus-5 "Review this design"
defapi text gemini --model gemini-3.8-flash "Write a haiku"
```

These identifiers are listed as active by Defapi, but have not been verified with
paid generation calls. Older models remain selectable with `--model`. See the
[2026-09-16 API review](docs/defapi-review-2026-09-16.md) for sources and limitations.

### Image

```sh
defapi image wan   --ratio 16:9 "A cat on a mountain"
defapi image mj    "Cyberpunk city --ar 16:9 --stylize 750"
defapi image gpt   --quality high "A logo for a coffee shop"
defapi image gpt2  --image https://example.com/photo.jpg "Make it night time"
defapi image google --model nano-banana-2 "Abstract art"
```

| Model    | Provider  | Notes                                      |
|----------|-----------|--------------------------------------------|
| `wan`    | Alibaba   | Wan 2.7, text-to-image                     |
| `mj`     | Midjourney| Imagine/edit, MJ params in prompt          |
| `gpt`    | OpenAI    | GPT-Image-1.5                              |
| `gpt2`   | OpenAI    | GPT-Image-2, supports `--image` for edits  |
| `google` | Google    | nano-banana-2 default, multiple models     |

Common flags: `--output/-o`, `--open`

### Video

```sh
defapi video seedance --duration 10 --ratio 16:9 "A dolphin jumping"
defapi video grok     --duration 15 "Aurora borealis timelapse"
defapi video sora     --variant sora-2-hd "Slow motion waterfall"
```

| Model      | Provider  | Durations       |
|------------|-----------|-----------------|
| `seedance` | ByteDance | 5, 10, 15s      |
| `grok`     | xAI       | 10, 15s         |
| `sora`     | OpenAI    | 10, 15, 25s     |

Common flags: `--duration`, `--ratio`, `--image` (image-to-video), `--open`

Videos save to `~/Downloads/defapi_<taskid>.mp4`. Images save to `~/Downloads/defapi_<taskid>.<ext>`.

Run `defapi <category> <model> --help` for full flag reference.
