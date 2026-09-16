# Working on defapi-cli

This is a small, standard-library-only Go CLI for Defapi text, image, and video
generation. Keep changes focused and preserve existing commands and aliases.

## Project map

- `main.go`: top-level command dispatch.
- `text.go`: provider-specific text requests, streaming, prompts, and text output.
- `image.go` and `video.go`: media flags and task submission.
- `api.go`: shared HTTP responses, task polling, downloads, and file opening.
- `main_test.go` and other `*_test.go` files: offline regression tests.
- `completions/`: Bash and Zsh completions; update both when flags/models change.
- `.github/workflows/`: CI and tagged release builds.
- `install.sh` and `Makefile`: installation and local builds.

## Development and verification

Use the Go version required by `go.mod`. Prefer the standard library over adding
dependencies. Format changed Go files with `gofmt`, then run:

```sh
go test ./...
go vet ./...
go build ./...
```

Add focused regression tests for behavior changes, using mock HTTP transports or
`httptest` instead of real generation calls. Check help output and synchronize the
README and shell completions for user-facing changes. Flags normally precede the
prompt; image commands also support output flags after the prompt.

## API integration

- The API base is `https://api.defapi.org`; credentials come from `DEFAPI_API_KEY`.
  Never print, commit, or embed credentials in fixtures.
- Verify endpoints, identifiers, payloads, and availability against current
  official Defapi documentation. Record dated findings and links in `docs/`.
  A catalog listing alone does not verify endpoint compatibility.
- Preserve defaults unless there is a verified migration reason. Text `--model`
  accepts arbitrary identifiers, so newly listed models usually need examples and
  completions rather than new commands. Do not enable “Coming Soon” models.
- Media submission is asynchronous: submit, poll `/api/task/query`, then download.
  Do not automatically retry paid submissions; a retry may create a second job.
- Handle HTTP, JSON, stream, and filesystem failures explicitly. Never report a
  failed or partial download as successfully saved, or destroy an existing output
  file before a replacement download completes.
- Keep automated checks offline. Only run paid generation when the user requests
  it. Do not install binaries globally or publish releases as part of routine fixes.

For website inspection, follow the user's browser preference and the installed
browser skill. This CLI has no browser UI of its own.
