package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchRegistryModels(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		models := []CommunityModel{
			{
				Name:        "test-model",
				Author:      "test-author",
				Description: "test description",
				URL:         "https://huggingface.co/test-author/test-model",
				Upvotes:     10,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	}))
	defer ts.Close()

	// Backup original URL and restore after test
	oldRegistryURL := RegistryURL
	RegistryURL = ts.URL
	defer func() { RegistryURL = oldRegistryURL }()

	models, err := fetchRegistryModels()
	if err != nil {
		t.Fatalf("fetchRegistryModels failed: %v", err)
	}

	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "test-model" {
		t.Errorf("expected model name test-model, got %s", models[0].Name)
	}
}

func TestFetchModelFeedback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock endpoint: /issues
		if strings.HasSuffix(r.URL.Path, "/issues") {
			issues := []GitHubIssue{
				{
					Number: 42,
					Title:  "Community Model: test-model",
					State:  "open",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(issues)
			return
		}

		// Mock endpoint: /issues/42/comments
		if strings.HasSuffix(r.URL.Path, "/issues/42/comments") {
			comments := []GitHubComment{
				{
					Body: "This is a great model!",
				},
			}
			comments[0].User.Login = "reviewer-bob"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(comments)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	oldGithubAPIBase := GithubAPIBase
	GithubAPIBase = ts.URL
	defer func() { GithubAPIBase = oldGithubAPIBase }()

	comments, err := fetchModelFeedback("test-model")
	if err != nil {
		t.Fatalf("fetchModelFeedback failed: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	if comments[0].Body != "This is a great model!" {
		t.Errorf("expected comment body to match, got %s", comments[0].Body)
	}
}
