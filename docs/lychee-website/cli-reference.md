# CLI Reference

The Lychee CLI is your primary interface for managing and running models locally.

## Core Commands

### `lychee serve`
Starts the Lychee API server and embedded dashboard.

### `lychee run <model>`
Runs a model. If the model is not present locally, it will automatically pull it.

### `lychee pull <model>`
Downloads a model from the registry without running it.

### `lychee list`
Lists all downloaded models and their sizes.

### `lychee rm <model>`
Deletes a locally stored model.

## Advanced Commands

### `lychee bench`
Runs the integrated performance benchmarking suite.
- `--model <name>`: Model to benchmark.
- `--epochs <N>`: Number of times to repeat the prompt.

### `lychee modelfile lint`
Statically analyzes a `Modelfile` for syntax errors or deprecated directives.

### `lychee hf pull <repo>`
Downloads a GGUF model directly from HuggingFace.
