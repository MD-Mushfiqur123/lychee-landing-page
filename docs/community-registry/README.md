# Lychee Community Registry

This is the companion repository that serves as the backend for the Lychee community showcase features.

## Setup Instructions

To activate the real CLI backend integration:

1. **Create a GitHub Repository**:
   Create a new public repository on GitHub named `community-registry` under your organization or personal account (e.g. `github.com/lychee-ai/community-registry`).

2. **Initialize Repository**:
   Push the contents of this folder (`docs/community-registry/`) to the default branch (`main`):
   ```bash
   git init
   git remote add origin https://github.com/lychee-ai/community-registry.git
   git branch -M main
   git add .
   git commit -m "Initial registry setup"
   git push -u origin main
   ```

3. **Configure Permissions**:
   Go to **Settings** > **Actions** > **General** > **Workflow permissions** in your GitHub repository settings and select **Read and write permissions**. This is required for the auto-merge workflow to append entries to `models/registry.json` and close the issues automatically.
