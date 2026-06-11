# Good First Issues

Copy and paste these templates into GitHub Issues to bootstrap the open-source community for Lychee.

---

### 1. Add syntax highlighting to VitePress docs for Modelfiles
**Description:** The new VitePress documentation site (`docs/lychee-website`) doesn't highlight `Modelfile` syntax. We need a custom prism/shiki grammar for Modelfile syntax (which is very similar to Dockerfile syntax).

### 2. Implement timeout support in Python SDK
**Description:** The Python SDK (`sdk/python`) currently uses default HTTP timeouts. Add a `timeout` parameter to the `Client` initialization and pass it to `httpx`.

### 3. Add `--json` flag to `lychee list`
**Description:** Add a flag to `cmd/cmd.go` for the `list` command to output the locally installed models in JSON format for easier machine parsing.

### 4. Create a VS Code snippet for Modelfiles
**Description:** Enhance the VS Code extension (`sdk/vscode`) to provide autocompletion snippets for common Modelfile directives like `FROM`, `SYSTEM`, `TEMPLATE`, and `PARAMETER`.

### 5. Add unit tests for `server/schema_validator.go`
**Description:** Write a comprehensive Go unit test suite for the JSON Schema validation logic to ensure nested objects and arrays are correctly enforced.

### 6. Document the `/api/embeddings` endpoint
**Description:** Add documentation to the VitePress site covering how to generate vector embeddings using Lychee.

### 7. Support `max_tokens` alias for `num_predict` in Python SDK
**Description:** For better OpenAI compatibility, allow users to pass `max_tokens` in the Python SDK kwargs, which should map to the `num_predict` option.

### 8. Add health check endpoint (`/api/health`)
**Description:** Implement a simple `GET /api/health` endpoint in `server/routes.go` that returns a 200 OK `{"status": "ok"}` when the engine is running.

### 9. Publish Rust SDK to crates.io
**Description:** The `lychee-rs` SDK is scaffolded but needs its GitHub Actions workflow configured to automatically publish tags to `crates.io`.

### 10. Add `lychee modelfile format` command
**Description:** Expand the new `lychee modelfile lint` command to include a `format` subcommand that automatically indents and formats a Modelfile, similar to `gofmt`.

### 11. Implement retry logic in JavaScript SDK
**Description:** Add automatic retry mechanisms with exponential backoff for `503 Service Unavailable` or network errors in the JS SDK.

### 12. Add Docker Compose example
**Description:** Create a `docker-compose.yml` file in the repository root demonstrating how to run Lychee alongside an application container (like a Python API).

### 13. Expose prefill tokens-per-second in CLI
**Description:** Update `cmd_run.go` to display the "Prompt eval rate" (prefill speed) alongside the generation speed when `--verbose` is passed.

### 14. Document the `LYCHEE_MAX_VRAM` env variable
**Description:** Add examples of how to strictly limit VRAM usage on multi-GPU setups in `docs/lychee-website/environment-variables.md`.

### 15. Create a Modelfile syntax guide
**Description:** Add a dedicated markdown page to the VitePress docs explaining each Modelfile instruction (`FROM`, `SYSTEM`, `LICENSE`, etc.).

### 16. Support custom system prompt via CLI flag
**Description:** Allow overriding a model's system prompt directly from the CLI during execution: `lychee run llama3 --system "You are a helpful assistant."`

### 17. Build a GitHub Action for Lychee
**Description:** Create a reusable GitHub Action that spins up a Lychee server in CI environments, allowing developers to run integration tests against local LLMs.

### 18. Add support for `.env` file loading
**Description:** Use the `godotenv` package to automatically load a `.env` file if it exists in the current directory, rather than relying solely on exported shell variables.

### 19. Create a Windows installer (MSI/NSIS)
**Description:** Convert the `scripts/install.ps1` script into a proper compiled Windows installer for a better user experience.

### 20. Implement context length truncation warning
**Description:** If a user sends a prompt that exceeds the model's configured context length, emit a clear warning log indicating that the context window is overflowing.
