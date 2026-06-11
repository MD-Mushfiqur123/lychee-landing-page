# Environment Variables

You can configure Lychee's behavior globally using environment variables.

| Variable | Default | Description |
|---|---|---|
| `LYCHEE_HOST` | `127.0.0.1:11434` | The IP and port the server binds to. |
| `LYCHEE_MODELS` | `~/.lychee/models` | Directory where model files are stored. |
| `LYCHEE_DEBUG` | `false` | Enables verbose logging and timing information. |
| `LYCHEE_NUM_PARALLEL` | `1` | Maximum number of concurrent generate requests processed in parallel. |
| `LYCHEE_MAX_VRAM` | `auto` | Hard limit on VRAM usage (in bytes). |
| `LYCHEE_DISABLE_H2C` | `false` | Fall back to HTTP/1.1 instead of H2C. |
| `LYCHEE_HF_TOKEN` | `""` | HuggingFace API token for downloading gated models. |
