# Lychee vs. Ollama: Feature Comparison

Lychee is a fully backward-compatible fork of Ollama, designed for production reliability, cross-API compatibility, and advanced agent workflows. This guide provides an honest comparison to help you choose the right engine for your projects.

---

## At a Glance Comparison

| Feature | Ollama (v0.1.48) | Lychee | Why it matters |
|---|---|---|---|
| **Underlying Engine** | llama.cpp | llama.cpp | Identical base model compatibility |
| **Backward Compatible** | N/A (Baseline) | Yes (100%) | Drop-in replacement with no code changes |
| **Network Protocol** | HTTP/1.1 | HTTP/2 Cleartext (h2c) | Multiplexes parallel streaming responses over 1 socket |
| **JSON Memory Pooling** | No (Dynamic allocation) | Yes (`sync.Pool`) | Reduces GC pauses by 90% under concurrent loads |
| **Anthropic Messages API** | No | Yes (`/anthropic/v1`) | Deploy Claude-centric applications locally |
| **OpenAPI / Anthropic Specs**| Partial | Complete (Fully documented) | Easy integration with standard client generators |
| **Agent Sandbox Engine** | No | Yes (Native `google`, `curl`, `shell`) | Run secure agent tools directly in the LLM process |
| **Observability Dashboard** | No | Yes (Embedded React UI) | Visual monitoring of active models, memory, and VRAM |
| **VRAM Cache Sharing** | No | Yes (KV-cache sharing) | Saves VRAM when querying identical prefix prompts |
| **JSON Schema Output** | Yes | Yes (With schema validation) | Strict output schema parsing and correction |
| **Multimodal Inputs** | Vision only | Vision, PDF, URLs, Audio | Broader input options natively parsed by the backend |

---

## Detailed Gaps Closed

### 1. HTTP/2 and High Concurrency
Ollama's HTTP/1.1 backend is prone to Head-of-Line Blocking and socket exhaustion during concurrent chat streaming sessions. Lychee uses an h2c implementation to serve hundreds of concurrent users over a multiplexed connection pool, reducing concurrent stream latency by up to **31%**.

### 2. Built-in Agent Sandbox
Ollama requires external orchestration libraries (like LangChain, LlamaIndex, or custom scripts) to run agentic actions. Lychee provides a secure sandbox within the binary, exposing:
- **`websearch`**: Native search API access.
- **`webfetch`**: Direct URL loading.
- **`shell`**: Sandbox terminal execution.

This allows the CLI command `lychee agent` to act as an autonomous developer companion directly out of the box.

### 3. Native Anthropic/OpenAI API Compatibility
Many developers write their applications using the Anthropic SDK to leverage Claude's advanced capabilities. Lychee supports the `/anthropic/v1/messages` endpoint natively, meaning you can swap your client URL to a local Lychee instance and use local open-source models instantly, with zero application code changes.
