# Migrating from Ollama

If you are already using Ollama, transitioning to Lychee is incredibly simple. Lychee is 100% compatible with the Ollama API, CLI commands, and Modelfiles.

## 1. Stop Ollama

Before starting Lychee, ensure Ollama is stopped so it releases port `11434`.

```bash
# Linux
sudo systemctl stop ollama

# macOS
launchctl stop com.ollama.ollama
```

## 2. Start Lychee

```bash
lychee serve
```

Lychee automatically binds to `localhost:11434` and will serve all your existing integrations (Open WebUI, LangChain, etc.) seamlessly.

## Key Differences
- **Performance**: Lychee uses H2C and `sync.Pool` for reduced latency.
- **HuggingFace**: Use `lychee hf pull <repo>` to fetch any GGUF directly.
- **Agents**: Use `lychee run <model> --experimental` to enable local shell tools.
