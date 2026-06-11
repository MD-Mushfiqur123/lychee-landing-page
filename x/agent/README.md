# 🕵️ Stateful Agent & Interactive Sandbox Approval System

The `x/agent` package provides the security sandbox and interactive TUI approval framework for Lychee's agentic execution mode. This allows local LLMs to invoke tools—such as running terminal commands or fetching remote URLs—safely, keeping the user in full control.

## How it works

When a model requests a tool call, Lychee intercept it and feeds it into the `ApprovalManager`. 

1. **Deny-List Check**: The command/arguments are compared against hard-coded dangerous patterns (`denyPatterns` e.g., fork bombs, recursive deletions like `rm -rf`, SSH key reads, etc.). If a match is found, the command is blocked instantly without prompting.
2. **Directory Boundary Check**: For filesystem actions, the sandbox verifies if the action falls outside the current working directory boundary (unless explicitly configured).
3. **Interactive TUI Prompt**: The user is presented with a clear terminal overlay showing:
   - The tool to be executed.
   - The specific parameters/commands.
   - A security warning (e.g. "Uses direct internet connection" or "Runs terminal commands").
   - Choices: **Allow once (y)**, **Always allow (a)**, **Deny (n)**, or **Custom reason / feedback**.
4. **Allowlist / Session Memory**: "Always allow" selections add hierarchical path prefixes or specific tool signatures to the session allowlist to prevent repetitive prompting.

## Built-In Tools

- **`bash`**: Runs terminal commands (restricted by the sandbox security constraints).
- **`web_search`**: Performs web search queries directly via Brave Search API or DuckDuckGo.
- **`web_fetch`**: Fetches and cleans HTML page text locally.

## CLI Usage

To run Lychee in agentic tool-execution mode:

```bash
lychee run gemma3 --experimental
```

### Controls inside Agent Prompt

- **/tools**: Lists all registered tools and active session approvals/permissions.
- **/clear**: Clears active session context and forgets all tool approvals.
- **Ctrl+O**: Expand/collapse the last tool execution output in the console.
