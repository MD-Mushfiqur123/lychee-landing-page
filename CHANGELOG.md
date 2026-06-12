# Changelog

All notable changes to Lychee, built as an extension of Ollama.

## [0.5.0-alpha] - 2026-06-12

### Added
- **Conversation Memory Store Enhancements**: Added pagination (`limit`/`offset`) and search query (`q`) parameters to `/api/conversations` conversation listing, automatic conversation title updating, and SQLite migration support.
- **Model Router Hardening**: Added health checking background loop, health-based endpoint resolution, a route status API endpoint (`/api/routes/:name/status`), active request tracking, weighted round-robin load balancing strategy, and a circuit breaker that trips after consecutive failures with recovery cooldown.
- **Structured Output Enhancements**: Added Server-Sent Events (SSE) streaming support for structured output generation retries, automatic temperature escalation (+0.1 per retry, capped at 1.2), and per-attempt timeout controls.
- **Pure Go SQLite Transition**: Replaced CGo-based `github.com/mattn/go-sqlite3` with pure-Go `modernc.org/sqlite` database driver to enable cross-compilation without C toolchains.
- **SDK Ecosystem & Tools**:
  - **Python SDK**: Structured packaging via `pyproject.toml` and version bump to `0.5.0a1`.
  - **JS SDK**: Renamed npm package to `lychee-js`, version bump to `0.5.0-alpha.1`.
  - **Rust SDK**: Added new async client endpoints matching the complete memory, router, structured output, and composer APIs, and integration test coverage.
  - **VSCode Extension**: Added commands to list models, run structured generation, and execute compose pipelines.
  - **CLI `compose` command**: Added CLI command to execute multi-model pipelines with live progress streaming.
- **CI/CD & Integration**: Transitioned build job to package-only compiles, enforced `CGO_ENABLED=0` in testing, created GitHub release automation workflow (`lychee-release.yml`), and generated test coverage reports.

### Added (Baseline Features vs Ollama)
- **Anthropic Messages API Adapter (`/v1/messages`)**: Run Claude-compatible workloads locally.
- **OpenAI Responses API (`/v1/responses`)**: Native support for structured JSON responses with JSON Schema validation.
- **Model Composer (`/api/compose`)**: Chain multiple models sequentially, substituting outputs from previous steps using customizable template variables (`{{input}}` and `{{step[n].output}}`).
- **Embedded Web Dashboard (`/dashboard/*`)**: A beautiful, embedded React-based web interface to chat with models, run configurations, and view system resources.
- **Stateful Agent Mode (`lychee run --experimental`)**: Run local agentic loops with TUI-driven sandbox approval gates for terminal commands, filesystem access, and web search.
- **Model Comparison (`lychee compare`)**: Benchmark multiple models head-to-head on the same prompt with side-by-side streaming output.
- **Rich Terminal Dashboard (`lychee stats --tui`)**: A live terminal-based monitor for running models, VRAM usage, and active context lengths.
- **Flexible Embeddings (`lychee embed`)**: Generate text embeddings with structured output support (JSON/CSV) and custom formatting.
- **Hardware Recommendations (`lychee scan`)**: Automatically detect system GPUs and VRAM capacity, mapping them to recommended models.
- **Client SDK Generator (`lychee generate-client`)**: Generate ready-to-run API client boilerplate code in Python, JavaScript, Rust, or Go for loaded models.
- **Direct HuggingFace Pull (`lychee hf pull`)**: Pull GGUF files directly from any HuggingFace repository with shard downloading and auto-resume.
- **GitHub Community Registry (`lychee community`)**: Explore and pull custom model configurations shared in the public repository index.

### Fixed
- Fixed error swallowing and unhandled database errors inside `memory_store.go` by adding structured logging.
- Fixed Windows compatibility in `openBrowser()`, resolving URL query-string execution bugs.
- Hardened `ensureServerRunning()`, replacing an infinite retry loop with a robust 15-second deadline.
- Improved model loading resource estimation in `scanHandler()` by correctly summing VRAM sizes of active models.
- Modularized CLI implementation by refactoring the monolithic `extras.go` into discrete command-specific files (`cmd_embed.go`, `cmd_scan.go`, etc.).

### Testing
- Added comprehensive unit testing suite across new CLI utilities (`cmd_embed_test.go`, `cmd_compare_test.go`, `cmd_generate_test.go`, and `community_registry_test.go`).
- Added full HTTP integration tests for all native APIs (`handler_integration_test.go`).
- Added extensive test coverage for model router health, circuit breakers, structured output timeouts, and memory pagination/search.
