# Changelog

All notable changes to Lychee, built as an extension of Ollama.

## [Unreleased] — Lychee vs Ollama Baseline

### Added (Features not in Ollama)
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
- **GitHub Community Registry (`lychee community`)**: Explore and pull custom models and configurations shared in the public repository index.

### Fixed (Bugs resolved compared to Ollama)
- Fixed Windows compatibility in `openBrowser()`, resolving URL query-string execution bugs.
- Hardened `ensureServerRunning()`, replacing an infinite retry loop with a robust 15-second deadline.
- Improved model loading resource estimation in `scanHandler()` by correctly summing VRAM sizes of active models.
- Modularized CLI implementation by refactoring the monolithic `extras.go` into discrete command-specific files (`cmd_embed.go`, `cmd_scan.go`, etc.).

### Testing
- Added comprehensive unit testing suite across new CLI utilities (`cmd_embed_test.go`, `cmd_compare_test.go`, `cmd_generate_test.go`, and `community_registry_test.go`).
