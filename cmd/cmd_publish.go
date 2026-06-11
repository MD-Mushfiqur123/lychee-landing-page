package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

// communityMockModels holds fallback sample data for the community showcase when offline.
var communityMockModels = []CommunityModel{
	{
		Name:        "llama3-coder-super",
		Author:      "dev_mind",
		Description: "Fine-tuned Llama 3 for extreme quality python completions.",
		URL:         "https://huggingface.co/dev_mind/llama3-coder-super",
		Upvotes:     42,
		SubmittedAt: time.Now().Add(-48 * time.Hour),
	},
	{
		Name:        "qwen-3b-uncensored",
		Author:      "mika_l",
		Description: "Uncensored assistant model with dynamic context scaling.",
		URL:         "https://huggingface.co/mika_l/qwen-3b-uncensored",
		Upvotes:     28,
		SubmittedAt: time.Now().Add(-7 * 24 * time.Hour),
	},
}

func NewCommunityPublishCmd() *cobra.Command {
	var author string
	var desc string
	var modelURL string

	cmd := &cobra.Command{
		Use:   "publish MODEL_NAME",
		Short: "Submit and share a customized model with the Lychee community showcase",
		Long:  `[PREVIEW] Opens the browser to submit a community model card issue.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]

			if author == "" {
				author = "anonymous"
			}
			if desc == "" {
				desc = "A community-shared local LLM integration."
			}
			if modelURL == "" {
				modelURL = "https://huggingface.co/models"
			}

			fmt.Fprintln(cmd.OutOrStdout(), "─────────────────────────────────────────────────────────────────")
			fmt.Fprintf(cmd.OutOrStdout(), "📦 Model:       %s\n", modelName)
			fmt.Fprintf(cmd.OutOrStdout(), "👤 Publisher:   @%s\n", author)
			fmt.Fprintf(cmd.OutOrStdout(), "📝 Description: %s\n", desc)
			fmt.Fprintf(cmd.OutOrStdout(), "🌐 URL:         %s\n", modelURL)
			fmt.Fprintln(cmd.OutOrStdout(), "─────────────────────────────────────────────────────────────────\n")

			body := fmt.Sprintf("**Author**: @%s\n**Description**: %s\n**URL**: %s", author, desc, modelURL)
			githubURL := fmt.Sprintf("https://github.com/lychee-ai/community-registry/issues/new?title=Community+Model%%3A+%s&body=%s",
				url.QueryEscape(modelName),
				url.QueryEscape(body),
			)

			fmt.Fprintln(cmd.OutOrStdout(), "📢 Redirecting to GitHub to submit model card...")
			err := openBrowser(githubURL)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Could not launch browser automatically. Please open this link manually to submit your model:\n\n%s\n", githubURL)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "✅ Browser opened! Please review and submit the issue to list your model.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&author, "author", "a", "", "Your developer handle (default: anonymous)")
	cmd.Flags().StringVarP(&desc, "description", "d", "", "Short description of the model fine-tune or configuration")
	cmd.Flags().StringVarP(&modelURL, "url", "u", "", "Public URL of the model weights repository (e.g. HuggingFace link)")

	return cmd
}

func NewCommunityListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trending models submitted by the Lychee community",
		Long:  `Displays live community model entries fetched from GitHub, with a fallback to local cache if offline.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "✨ Lychee Community Showcase - Live Models:")
			fmt.Fprintln(cmd.OutOrStdout(), "Fetching latest registry entries from GitHub...")

			models, err := fetchRegistryModels()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "\n⚠️  Failed to fetch live community registry: %v\n", err)
				fmt.Fprintln(cmd.OutOrStdout(), "   Showing local cached models instead.\n")
				models = communityMockModels
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "✅ Successfully loaded live community models.\n")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "────────────────────────────────────────────────────────────────────────────────────────────────")
			fmt.Fprintf(cmd.OutOrStdout(), "  %-24s %-12s %-8s %-45s\n", "MODEL NAME", "AUTHOR", "UPVOTES", "REPOSITORY LINK")
			fmt.Fprintln(cmd.OutOrStdout(), "────────────────────────────────────────────────────────────────────────────────────────────────")

			for _, m := range models {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-24s @%-11s %-8d %-45s\n", m.Name, m.Author, m.Upvotes, m.URL)
				fmt.Fprintf(cmd.OutOrStdout(), "    \"%s\"\n\n", m.Description)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "────────────────────────────────────────────────────────────────────────────────────────────────")
			fmt.Fprintln(cmd.OutOrStdout(), "To pull any model, use: lychee pull <repository_link>")
			return nil
		},
	}
	return cmd
}

// WireCommunityPublishCommands attaches publish and list as subcommands of the given parent.
func WireCommunityPublishCommands(parent *cobra.Command) {
	parent.AddCommand(NewCommunityPublishCmd())
	parent.AddCommand(NewCommunityListCmd())
}
