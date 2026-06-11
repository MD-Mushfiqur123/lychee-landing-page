# 🍒 Lychee

### Universal Local Inference, Prompts Caching, and Stateful Agent Sandbox

Lychee is a state-of-the-art local LLM runtime, designed to be the ultimate drop-in upgrade for your AI development workflow. While keeping full compatibility with the existing open-model ecosystem, Lychee extends local inference with enterprise-grade capabilities.

<p align="center">
  <a href="https://github.com/lychee/lychee">
    <img src="https://img.shields.io/github/license/lychee/lychee?style=flat-square&color=A51C30" alt="License"/>
  </a>
  <a href="https://github.com/lychee/lychee/pkgs/container/lychee">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue?style=flat-square" alt="GHCR Docker Image"/>
  </a>
</p>

---

## ⚡ The 5 Moats: Why Lychee?

Lychee goes far beyond standard model running by providing features that typically require heavy cloud orchestration, all running locally on your hardware:

1. **Universal API Gateway**
   Lychee is the only local model runner that natively speaks multiple API protocols. Run your existing codebases unmodified, whether they are built for **OpenAI (Completions)**, **Anthropic (Messages)**, or the **OpenAI Responses API**.
   
2. **Native Agent Mode**
   An embedded stateful agent execution framework (`x/agent/`) built directly into the CLI and API. Run agents with sandbox approval gates for file access, web search, and terminal command execution.

3. **Prompt Caching & KV Sharing**
   Drastically reduce Time to First Token (TTFT) by automatically caching prompt prefixes. Built on top of the upstream Ollama scheduler, Lychee leverages message prefix-hash matching to optimize cache reuse during multi-turn chats.

4. **Model Composer**
   Chain multiple local models together sequentially. Run complex pipelines where the output of one model feeds directly into the prompt template (`{{input}}` / `{{step[n].output}}`) of subsequent models—perfect for local classification, translation, or structured processing chains.

5. **Embedded Web Dashboard**
   Lychee comes with a built-in, fully-featured React-based SPA console served directly from the server at `/dashboard/`. Interact with models, manage configurations, view system resource utilization, and test API endpoints—all out-of-the-box.

---

## 🔌 API Compatibility Layer

Lychee acts as a local proxy that translates industry-standard APIs into optimized backend executions.

| Endpoint | Standard | Features Supported | Status |
| --- | --- | --- | --- |
| `/api/chat` | Lychee Native | Streaming, Tools, Multi-turn Chat | ✅ Production |
| `/api/generate` | Lychee Native | Streaming, Format, Options | ✅ Production |
| `/v1/chat/completions` | OpenAI | Chat completions, System prompts | ✅ Production |
| `/v1/messages` | Anthropic | Messages API, Streaming, System prompt | ✅ Production |
| `/v1/responses` | OpenAI Responses | Structured responses, validation | ✅ Production |

---

## 📦 Installation

> [!NOTE]
> Lychee is currently in active release candidate phase.

### Build from Source
```bash
git clone https://github.com/lychee/lychee.git
cd lychee
go build -o lychee .
sudo mv lychee /usr/local/bin/
```

### Advanced Installation (Requires GitHub Releases)
The installation scripts below retrieve pre-built binaries. Note: These require a published release.
* **macOS & Linux:**
  ```bash
  curl -fsSL https://raw.githubusercontent.com/lychee/lychee/main/scripts/install.sh | sh
  ```
* **Windows (PowerShell):**
  ```powershell
  irm https://raw.githubusercontent.com/lychee/lychee/main/scripts/install.ps1 | iex
  ```

### Docker (Coming Soon)
> Docker images will be published to GHCR once the first stable release is tagged.
> The Dockerfile is ready in the repository root — you can build locally:
> ```bash
> docker build -t lychee .
> docker run -d -v lychee:/root/.lychee -p 11434:11434 --name lychee lychee
> ```

---

## 💡 Core Features Showcase

### 1. Direct HuggingFace Pull (`lychee hf`)
Search and download any GGUF format model directly from HuggingFace Hub with resume support, concurrent multi-shard downloads, and SHA256 integrity verification:
```bash
# Search catalog models
lychee hf search llama

# Download the recommended quantization of a HuggingFace model
lychee hf pull microsoft/Phi-3-mini-4k-instruct-gguf

# List all available quantizations for a model repository
lychee hf pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --list

# Download a specific quantization variant
lychee hf pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --quant q5_k_m
```

### 2. Drop-in Anthropic SDK Support (`/v1/messages`)
Deploy Claude-compatible workflows locally. Point the official Anthropic client to Lychee:
```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:11434/v1",
    api_key="lychee-local"
)

message = client.messages.create(
    model="gemma3",
    max_tokens=1000,
    messages=[
        {"role": "user", "content": "Explain quantum computing in one sentence."}
    ]
)
print(message.content[0].text)
```

### 3. OpenAI Responses API (`/v1/responses`)
Generate structured data conforming to a JSON Schema directly from local models:
```bash
curl http://localhost:11434/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemma3",
    "prompt": "Extract information: Jane Doe is a 28-year-old engineer.",
    "response_format": {
      "type": "json_schema",
      "json_schema": {
        "name": "profile",
        "schema": {
          "type": "object",
          "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer"},
            "occupation": {"type": "string"}
          },
          "required": ["name", "age", "occupation"]
        }
      }
    }
  }'
```

### 4. Interactive Agent Sandbox (`lychee run --experimental`)
Run LLM-driven agents with local tool support. Lychee prompts you with terminal UI approval before executing commands, searching the web, or accessing the filesystem:
```bash
# Start an interactive agent loop with bash, web search, and web fetch capabilities
lychee run gemma3 --experimental

# Inside the prompt, ask:
# >>> Search the web for latest Rust news and compile a list of highlights in rust_news.txt
#
# Lychee will prompt you in real time to approve/reject the web search and bash write actions!
```

### 5. Multi-Model Benchmarking & Comparison (`lychee compare`)
Evaluate multiple local models head-to-head on the same prompt with real-time, side-by-side streaming output and response metrics (tokens/sec, time-to-first-token):
```bash
# Compare gemma3 and phi3 side-by-side
lychee compare "Explain quantum computing in three sentences." gemma3 phi3
```

### 6. Interactive Terminal Performance Dashboard (`lychee stats --tui`)
Monitor active models, VRAM usage, token throughput, and model context usage in real time, directly from your shell, using a beautiful Bubbletea-powered terminal UI:
```bash
lychee stats --tui
```

### 7. Hardware Detection & Code Generation
Optimizing and integrating local models has never been simpler:
* **Hardware Scan (`lychee scan`)**: Scan your hardware configuration and receive personalized local model sizing and execution recommendations.
* **Code Boilerplate Generator (`lychee generate-client`)**: Generate ready-to-run, native API client code in Python, JavaScript, Rust, or Go for any loaded model.

---

## 🚀 Quickstart

1. **Start the Lychee server**
   ```bash
   lychee serve
   ```

2. **Run your first model**
   ```bash
   lychee run gemma3
   ```

3. **Interact via API**
   
   *Using the Anthropic Messages API:*
   ```bash
   curl http://localhost:11434/v1/messages \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gemma3",
       "messages": [{"role": "user", "content": "Hello!"}],
       "max_tokens": 1024
     }'
   ```

---

## 🛠️ Supported Backends

Lychee leverages the best-in-class local inference engines under the hood:
* **llama.cpp** — optimized CPU/GPU execution for Apple Silicon (Metal), NVIDIA (CUDA), and AMD (ROCm).
* **MLX** — native Apple Silicon machine learning framework for maximum performance on macOS.

---

## 📄 License

Lychee is open-source software licensed under the [MIT License](LICENSE).
