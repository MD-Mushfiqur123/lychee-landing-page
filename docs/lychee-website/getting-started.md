# Getting Started with Lychee

Lychee is a powerful, high-performance local LLM engine. It is a hard fork of Ollama, designed for advanced use cases like Agentic execution, multi-SDK compatibility, and high concurrency.

## Installation

### macOS and Linux
```bash
curl -fsSL https://raw.githubusercontent.com/lychee/lychee/main/scripts/install.sh | sh
```

### Windows
```powershell
irm https://raw.githubusercontent.com/lychee/lychee/main/scripts/install.ps1 | iex
```

## Running Your First Model

```bash
lychee run llama3
```

This will automatically pull the model from the Lychee registry (or HuggingFace) and drop you into an interactive chat prompt.
