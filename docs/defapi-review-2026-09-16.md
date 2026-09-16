# Defapi compatibility review — 2026-09-16

## Catalog changes

The public catalog lists these newer text models as active:

| CLI | Identifier | Source |
| --- | --- | --- |
| `text gpt` | `openai/gpt-5.6-sol` | [Defapi](https://defapi.org/model/openai/gpt-5.6-sol) |
| `text claude` | `anthropic/claude-opus-5` | [Defapi](https://defapi.org/model/anthropic/claude-opus-5) |
| `text gemini` | `google/gemini-3.8-flash` | [Defapi](https://defapi.org/model/google/gemini-3.8-flash) |

All text commands accept arbitrary `--model` values. These newer models are now
the defaults, as requested by the user. Help text, README examples, and Bash/Zsh
completions match these defaults; older models remain selectable with `--model`.
Availability in the catalog does not establish compatibility with every endpoint
or parameter. In particular, the GPT-5.6 Sol page says API documentation is not
available. No authenticated generation requests were made during this review.

The [homepage](https://defapi.org/) lists GPT-Image-2.5 Sunburst and Flare as
**Coming Soon**. Do not add them to the supported image model switch yet.

## Gemini request corrections

The existing CLI uses the Gemini-compatible `generateContent` API. Updated its
request construction according to Google's [API reference](https://ai.google.dev/api/generate-content):

- Send `--system` as `systemInstruction`, separate from the user prompt.
- Forward `--max-tokens`, `--temperature`, and `--top-p` through `generationConfig`
  as `maxOutputTokens`, `temperature`, and `topP`; they were previously ignored.
- Request `alt=sse` and an SSE Accept header for streaming, matching the CLI parser.

These changes are tested offline. Defapi's handling of these native Gemini fields
still needs an authenticated smoke test before claiming end-to-end verification.

## Remaining investigation

No verified media endpoint migration was established in this review. The public
model descriptions are not a complete versioned API changelog, and the browser
connection was unavailable for interactive API-tab inspection. Keep existing
image/video request schemas until current Defapi API schemas can be checked.

Future work: bounded polling/network timeouts and explicit provider-error handling
inside SSE streams. Those need deliberate timeout and partial-output semantics,
so they are outside this focused fix pass.
