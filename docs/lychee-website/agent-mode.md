# Agent Mode

Lychee introduces a native interactive Agent environment. The LLM can dynamically write shell commands, fetch web pages, and perform multi-step reasoning.

## Starting an Agent

Use the standard `run` command with the `--experimental` or agent flags.

```bash
lychee run llama3-toolcall --agent
```

## Built-in Tools

- `web_search`: Fetches DuckDuckGo results.
- `web_fetch`: Reads markdown content from a URL.
- `shell_execute`: Runs arbitrary bash/powershell commands (sandboxing must be configured).

## Example Interaction

```text
> Find the latest release of Kubernetes and summarize the changelog.

[Agent: Calling web_search("Kubernetes latest release")]
[Agent: Calling web_fetch("https://github.com/kubernetes/kubernetes/releases/latest")]

The latest release of Kubernetes is v1.30.0. The major changes include...
```
