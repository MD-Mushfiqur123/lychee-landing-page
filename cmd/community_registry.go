package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var (
	// RegistryURL is the raw JSON URL for listing community models.
	RegistryURL = "https://raw.githubusercontent.com/lychee-ai/community-registry/main/models/registry.json"

	// GithubAPIBase is the base URL for community-registry repository API.
	GithubAPIBase = "https://api.github.com/repos/lychee-ai/community-registry"
)

// CommunityModel represents a community-submitted model entry.
type CommunityModel struct {
	Name        string    `json:"name"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Upvotes     int       `json:"upvotes"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// GitHubIssue represents a simplified GitHub Issue struct.
type GitHubIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// GitHubComment represents a simplified GitHub Comment struct.
type GitHubComment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// fetchRegistryModels retrieves models from the community-registry repository raw JSON.
func fetchRegistryModels() ([]CommunityModel, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", RegistryURL, nil)
	if err != nil {
		return nil, err
	}

	// Optional GitHub token to avoid rate limits
	if token := os.Getenv("LYCHEE_GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned status code %d", resp.StatusCode)
	}

	var models []CommunityModel
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, err
	}

	return models, nil
}

// fetchModelFeedback fetches reviews/comments for a specific model ID from GitHub Issues.
func fetchModelFeedback(modelID string) ([]GitHubComment, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Fetch issues labeled "model-card"
	url := fmt.Sprintf("%s/issues?labels=model-card&state=open&per_page=100", GithubAPIBase)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("LYCHEE_GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned status code %d", resp.StatusCode)
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	issueNumber := -1
	for _, issue := range issues {
		if issue.Title == modelID ||
			issue.Title == "Community Model: "+modelID ||
			fmt.Sprintf("Community Model: %s", modelID) == issue.Title {
			issueNumber = issue.Number
			break
		}
	}

	if issueNumber == -1 {
		return nil, fmt.Errorf("no open community model card found for ID %q", modelID)
	}

	// Fetch comments for that issue
	commentsURL := fmt.Sprintf("%s/issues/%d/comments?per_page=100", GithubAPIBase, issueNumber)
	reqComments, err := http.NewRequest("GET", commentsURL, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("LYCHEE_GITHUB_TOKEN"); token != "" {
		reqComments.Header.Set("Authorization", "token "+token)
	}

	respComments, err := client.Do(reqComments)
	if err != nil {
		return nil, err
	}
	defer respComments.Body.Close()

	if respComments.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned status code %d when fetching comments", respComments.StatusCode)
	}

	var comments []GitHubComment
	if err := json.NewDecoder(respComments.Body).Decode(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// openBrowser opens the specified URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
