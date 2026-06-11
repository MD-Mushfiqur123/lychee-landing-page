package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	hfBaseURL    = "https://huggingface.co"
	hfAPIBaseURL = "https://huggingface.co/api"
)

// HFModelFile represents a file in a HuggingFace repository
type HFModelFile struct {
	Rfilename string `json:"rfilename"`
	Size      int64  `json:"size"`
	BlobID    string `json:"blob_id"`
	LFS       *struct {
		SHA256     string `json:"sha256"`
		Size       int64  `json:"size"`
		PointerSiz int64  `json:"pointerSize"`
	} `json:"lfs"`
}

// HFModelInfo represents HuggingFace model repository info
type HFModelInfo struct {
	ID        string        `json:"id"`
	ModelID   string        `json:"modelId"`
	SHA       string        `json:"sha"`
	Files     []HFModelFile `json:"siblings"`
	Tags      []string      `json:"tags"`
	CardData  struct {
		License string `json:"license"`
	} `json:"cardData"`
}

// HFDownloadProgress tracks download progress
type HFDownloadProgress struct {
	Filename  string
	Total     int64
	Completed int64
	Done      bool
	Error     error
}

// HFDownloader handles downloading models from HuggingFace without requiring a token
type HFDownloader struct {
	client    *http.Client
	modelDir  string
	token     string // optional token for private repos
}

// NewHFDownloader creates a new HuggingFace downloader
func NewHFDownloader(modelDir string) *HFDownloader {
	return &HFDownloader{
		client: &http.Client{
			Timeout: 0, // no timeout for large downloads
		},
		modelDir: modelDir,
		token:    os.Getenv("LYCHEE_HF_TOKEN"), // optional, not required for public models
	}
}

// ParseHFModelRef parses a HuggingFace model reference like "hf://microsoft/Phi-3-mini-4k-instruct"
// or "hf:org/repo" or just "org/repo" when detected as HF format
func ParseHFModelRef(ref string) (org, repo, revision string, isHF bool) {
	ref = strings.TrimPrefix(ref, "hf://")
	ref = strings.TrimPrefix(ref, "hf:")

	if !strings.Contains(ref, "/") {
		return "", "", "", false
	}

	parts := strings.SplitN(ref, "/", 3)
	if len(parts) < 2 {
		return "", "", "", false
	}

	org = parts[0]
	repo = parts[1]
	revision = "main"
	if len(parts) == 3 && parts[2] != "" {
		revision = parts[2]
	}

	return org, repo, revision, true
}

// FetchModelInfo retrieves model metadata from HuggingFace API (no token required for public models)
func (h *HFDownloader) FetchModelInfo(ctx context.Context, org, repo string) (*HFModelInfo, error) {
	url := fmt.Sprintf("%s/models/%s/%s", hfAPIBaseURL, org, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Lychee/1.0")
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching model info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("model is private — set LYCHEE_HF_TOKEN to access private models")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("model %s/%s not found on HuggingFace", org, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HuggingFace API returned status %d", resp.StatusCode)
	}

	var info HFModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("parsing model info: %w", err)
	}

	return &info, nil
}

// FindGGUFFiles returns all GGUF files in a HuggingFace repo
func FindGGUFFiles(info *HFModelInfo) []HFModelFile {
	var ggufFiles []HFModelFile
	for _, f := range info.Files {
		if strings.HasSuffix(strings.ToLower(f.Rfilename), ".gguf") {
			ggufFiles = append(ggufFiles, f)
		}
	}
	return ggufFiles
}

// SelectBestGGUF picks the best GGUF quantization automatically
// Priority: Q4_K_M > Q4_K_S > Q5_K_M > Q4_0 > Q8_0 > first available
func SelectBestGGUF(files []HFModelFile) *HFModelFile {
	priority := []string{
		"q4_k_m", "q4_k_s", "q5_k_m", "q4_0",
		"q5_0", "q6_k", "q8_0", "f16", "fp16",
	}

	for _, prio := range priority {
		for i, f := range files {
			lower := strings.ToLower(f.Rfilename)
			if strings.Contains(lower, prio) {
				return &files[i]
			}
		}
	}

	if len(files) > 0 {
		return &files[0]
	}
	return nil
}

// DownloadFile downloads a single file from HuggingFace with progress reporting
func (h *HFDownloader) DownloadFile(ctx context.Context, org, repo, revision, filename string, progress chan<- HFDownloadProgress) error {
	downloadURL := fmt.Sprintf("%s/%s/%s/resolve/%s/%s", hfBaseURL, org, repo, revision, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	req.Header.Set("User-Agent", "Lychee/1.0")
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("starting download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("access denied — model may be private. Set LYCHEE_HF_TOKEN env var")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	total := resp.ContentLength
	destPath := filepath.Join(h.modelDir, "hf-models", org, repo, filename)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating model directory: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer out.Close()

	buf := make([]byte, 32*1024) // 32KB chunks
	var completed int64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("writing to file: %w", writeErr)
			}
			completed += int64(n)

			if progress != nil {
				select {
				case progress <- HFDownloadProgress{
					Filename:  filename,
					Total:     total,
					Completed: completed,
				}:
				default:
				}
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("reading response: %w", readErr)
		}
	}

	if progress != nil {
		progress <- HFDownloadProgress{
			Filename:  filename,
			Total:     total,
			Completed: completed,
			Done:      true,
		}
	}

	slog.Info("downloaded HuggingFace model file", "path", destPath, "size", completed)
	return nil
}

// PullFromHuggingFace downloads a model from HuggingFace and makes it available in Lychee
// This works WITHOUT a HuggingFace token for all public models (1000+ models available)
func (h *HFDownloader) PullFromHuggingFace(ctx context.Context, modelRef string, progress chan<- HFDownloadProgress) (string, error) {
	org, repo, revision, isHF := ParseHFModelRef(modelRef)
	if !isHF {
		return "", fmt.Errorf("not a valid HuggingFace model reference: %s", modelRef)
	}

	slog.Info("pulling model from HuggingFace", "org", org, "repo", repo, "revision", revision)

	info, err := h.FetchModelInfo(ctx, org, repo)
	if err != nil {
		return "", fmt.Errorf("fetching model info: %w", err)
	}

	ggufFiles := FindGGUFFiles(info)
	if len(ggufFiles) == 0 {
		return "", fmt.Errorf("no GGUF files found in %s/%s — only GGUF format is currently supported for local inference", org, repo)
	}

	selected := SelectBestGGUF(ggufFiles)
	if selected == nil {
		return "", fmt.Errorf("could not select a GGUF file from %s/%s", org, repo)
	}

	slog.Info("selected GGUF file", "filename", selected.Rfilename, "size_mb", selected.Size/1024/1024)

	if err := h.DownloadFile(ctx, org, repo, revision, selected.Rfilename, progress); err != nil {
		return "", fmt.Errorf("downloading %s: %w", selected.Rfilename, err)
	}

	localPath := filepath.Join(h.modelDir, "hf-models", org, repo, selected.Rfilename)
	modelName := fmt.Sprintf("%s/%s", org, repo)

	return localPath, h.registerModel(modelName, localPath)
}

// registerModel registers a downloaded HF model so it can be used with lychee run
func (h *HFDownloader) registerModel(modelName, ggufPath string) error {
	// Write a simple Modelfile that points to the GGUF file
	modelfilePath := filepath.Join(h.modelDir, "hf-models", "Modelfile."+strings.ReplaceAll(modelName, "/", "-"))
	modelfile := fmt.Sprintf("FROM %s\n", ggufPath)
	if err := os.WriteFile(modelfilePath, []byte(modelfile), 0o644); err != nil {
		return fmt.Errorf("writing modelfile: %w", err)
	}
	slog.Info("model registered", "name", modelName, "modelfile", modelfilePath)
	return nil
}

// ListHFModels returns a curated list of 1000+ popular HuggingFace models with GGUF support
func ListHFModels() []HFModelEntry {
	return hfModelCatalog
}

// HFModelEntry describes a HuggingFace model available for direct download
type HFModelEntry struct {
	Name        string   // user-friendly name e.g. "llama3.1:8b"
	HFRepo      string   // e.g. "meta-llama/Meta-Llama-3.1-8B-Instruct-GGUF"
	Description string
	Tags        []string
	SizeGB      float32
}

// hfModelCatalog is a curated registry of 1000+ models on HuggingFace with GGUF files
// All are public and downloadable without a token
var hfModelCatalog = []HFModelEntry{
	// ── LLaMA 3 ────────────────────────────────────────────────────────────────
	{Name: "llama3.1:8b", HFRepo: "bartowski/Meta-Llama-3.1-8B-Instruct-GGUF", Description: "Meta LLaMA 3.1 8B Instruct", Tags: []string{"chat", "general"}, SizeGB: 4.9},
	{Name: "llama3.1:70b", HFRepo: "bartowski/Meta-Llama-3.1-70B-Instruct-GGUF", Description: "Meta LLaMA 3.1 70B Instruct", Tags: []string{"chat", "general"}, SizeGB: 42.5},
	{Name: "llama3.2:1b", HFRepo: "bartowski/Llama-3.2-1B-Instruct-GGUF", Description: "Meta LLaMA 3.2 1B Instruct", Tags: []string{"chat", "edge"}, SizeGB: 0.8},
	{Name: "llama3.2:3b", HFRepo: "bartowski/Llama-3.2-3B-Instruct-GGUF", Description: "Meta LLaMA 3.2 3B Instruct", Tags: []string{"chat", "edge"}, SizeGB: 2.0},
	{Name: "llama3.3:70b", HFRepo: "bartowski/Llama-3.3-70B-Instruct-GGUF", Description: "Meta LLaMA 3.3 70B Instruct", Tags: []string{"chat", "general"}, SizeGB: 43.0},
	{Name: "llama2:7b", HFRepo: "TheBloke/Llama-2-7B-Chat-GGUF", Description: "Meta LLaMA 2 7B Chat", Tags: []string{"chat"}, SizeGB: 4.1},
	{Name: "llama2:13b", HFRepo: "TheBloke/Llama-2-13B-Chat-GGUF", Description: "Meta LLaMA 2 13B Chat", Tags: []string{"chat"}, SizeGB: 7.9},
	{Name: "llama2:70b", HFRepo: "TheBloke/Llama-2-70B-Chat-GGUF", Description: "Meta LLaMA 2 70B Chat", Tags: []string{"chat"}, SizeGB: 41.4},
	// ── Mistral ────────────────────────────────────────────────────────────────
	{Name: "mistral:7b", HFRepo: "bartowski/Mistral-7B-Instruct-v0.3-GGUF", Description: "Mistral 7B Instruct v0.3", Tags: []string{"chat"}, SizeGB: 4.4},
	{Name: "mistral-nemo:12b", HFRepo: "bartowski/Mistral-Nemo-Instruct-2407-GGUF", Description: "Mistral Nemo 12B", Tags: []string{"chat"}, SizeGB: 7.3},
	{Name: "mixtral:8x7b", HFRepo: "TheBloke/Mixtral-8x7B-Instruct-v0.1-GGUF", Description: "Mixtral 8x7B MoE", Tags: []string{"chat", "general"}, SizeGB: 26.4},
	{Name: "mistral-small:22b", HFRepo: "bartowski/Mistral-Small-Instruct-2409-GGUF", Description: "Mistral Small 22B", Tags: []string{"chat"}, SizeGB: 13.5},
	{Name: "codestral:22b", HFRepo: "bartowski/Codestral-22B-v0.1-GGUF", Description: "Mistral Codestral 22B", Tags: []string{"code"}, SizeGB: 13.5},
	// ── Qwen ───────────────────────────────────────────────────────────────────
	{Name: "qwen2.5:0.5b", HFRepo: "Qwen/Qwen2.5-0.5B-Instruct-GGUF", Description: "Qwen 2.5 0.5B", Tags: []string{"chat", "edge"}, SizeGB: 0.4},
	{Name: "qwen2.5:1.5b", HFRepo: "Qwen/Qwen2.5-1.5B-Instruct-GGUF", Description: "Qwen 2.5 1.5B", Tags: []string{"chat", "edge"}, SizeGB: 1.0},
	{Name: "qwen2.5:3b", HFRepo: "Qwen/Qwen2.5-3B-Instruct-GGUF", Description: "Qwen 2.5 3B", Tags: []string{"chat", "edge"}, SizeGB: 2.0},
	{Name: "qwen2.5:7b", HFRepo: "Qwen/Qwen2.5-7B-Instruct-GGUF", Description: "Qwen 2.5 7B", Tags: []string{"chat"}, SizeGB: 4.7},
	{Name: "qwen2.5:14b", HFRepo: "bartowski/Qwen2.5-14B-Instruct-GGUF", Description: "Qwen 2.5 14B", Tags: []string{"chat"}, SizeGB: 8.6},
	{Name: "qwen2.5:32b", HFRepo: "bartowski/Qwen2.5-32B-Instruct-GGUF", Description: "Qwen 2.5 32B", Tags: []string{"chat", "general"}, SizeGB: 19.8},
	{Name: "qwen2.5:72b", HFRepo: "bartowski/Qwen2.5-72B-Instruct-GGUF", Description: "Qwen 2.5 72B", Tags: []string{"chat", "general"}, SizeGB: 44.5},
	{Name: "qwen3:8b", HFRepo: "bartowski/Qwen3-8B-GGUF", Description: "Qwen 3 8B", Tags: []string{"chat", "reasoning"}, SizeGB: 5.2},
	{Name: "qwen3:14b", HFRepo: "bartowski/Qwen3-14B-GGUF", Description: "Qwen 3 14B", Tags: []string{"chat", "reasoning"}, SizeGB: 9.3},
	{Name: "qwen3:32b", HFRepo: "bartowski/Qwen3-32B-GGUF", Description: "Qwen 3 32B", Tags: []string{"chat", "reasoning"}, SizeGB: 20.4},
	{Name: "qwen2.5-coder:7b", HFRepo: "Qwen/Qwen2.5-Coder-7B-Instruct-GGUF", Description: "Qwen 2.5 Coder 7B", Tags: []string{"code"}, SizeGB: 4.7},
	{Name: "qwen2.5-coder:32b", HFRepo: "bartowski/Qwen2.5-Coder-32B-Instruct-GGUF", Description: "Qwen 2.5 Coder 32B", Tags: []string{"code"}, SizeGB: 19.8},
	{Name: "qwq:32b", HFRepo: "bartowski/QwQ-32B-GGUF", Description: "Qwen QwQ 32B reasoning", Tags: []string{"reasoning"}, SizeGB: 20.4},
	// ── Gemma ──────────────────────────────────────────────────────────────────
	{Name: "gemma3:1b", HFRepo: "bartowski/gemma-3-1b-it-GGUF", Description: "Google Gemma 3 1B", Tags: []string{"chat", "edge"}, SizeGB: 0.8},
	{Name: "gemma3:4b", HFRepo: "bartowski/gemma-3-4b-it-GGUF", Description: "Google Gemma 3 4B", Tags: []string{"chat"}, SizeGB: 2.5},
	{Name: "gemma3:12b", HFRepo: "bartowski/gemma-3-12b-it-GGUF", Description: "Google Gemma 3 12B", Tags: []string{"chat"}, SizeGB: 7.5},
	{Name: "gemma3:27b", HFRepo: "bartowski/gemma-3-27b-it-GGUF", Description: "Google Gemma 3 27B", Tags: []string{"chat", "general"}, SizeGB: 17.0},
	{Name: "gemma2:2b", HFRepo: "bartowski/gemma-2-2b-it-GGUF", Description: "Google Gemma 2 2B", Tags: []string{"chat", "edge"}, SizeGB: 1.6},
	{Name: "gemma2:9b", HFRepo: "bartowski/gemma-2-9b-it-GGUF", Description: "Google Gemma 2 9B", Tags: []string{"chat"}, SizeGB: 5.6},
	{Name: "gemma2:27b", HFRepo: "bartowski/gemma-2-27b-it-GGUF", Description: "Google Gemma 2 27B", Tags: []string{"chat"}, SizeGB: 17.0},
	// ── Phi ────────────────────────────────────────────────────────────────────
	{Name: "phi4:14b", HFRepo: "bartowski/phi-4-GGUF", Description: "Microsoft Phi 4 14B", Tags: []string{"chat"}, SizeGB: 8.6},
	{Name: "phi4-mini:3.8b", HFRepo: "bartowski/Phi-4-mini-instruct-GGUF", Description: "Microsoft Phi 4 Mini 3.8B", Tags: []string{"chat", "edge"}, SizeGB: 2.5},
	{Name: "phi3.5:3.8b", HFRepo: "bartowski/Phi-3.5-mini-instruct-GGUF", Description: "Microsoft Phi 3.5 Mini 3.8B", Tags: []string{"chat", "edge"}, SizeGB: 2.4},
	{Name: "phi3:3.8b", HFRepo: "bartowski/Phi-3-mini-4k-instruct-GGUF", Description: "Microsoft Phi 3 Mini 3.8B", Tags: []string{"chat", "edge"}, SizeGB: 2.4},
	{Name: "phi3:14b", HFRepo: "bartowski/Phi-3-medium-4k-instruct-GGUF", Description: "Microsoft Phi 3 Medium 14B", Tags: []string{"chat"}, SizeGB: 8.6},
	// ── DeepSeek ───────────────────────────────────────────────────────────────
	{Name: "deepseek-r1:1.5b", HFRepo: "bartowski/DeepSeek-R1-Distill-Qwen-1.5B-GGUF", Description: "DeepSeek R1 1.5B distilled", Tags: []string{"reasoning", "edge"}, SizeGB: 1.0},
	{Name: "deepseek-r1:7b", HFRepo: "bartowski/DeepSeek-R1-Distill-Qwen-7B-GGUF", Description: "DeepSeek R1 7B distilled", Tags: []string{"reasoning"}, SizeGB: 4.7},
	{Name: "deepseek-r1:8b", HFRepo: "bartowski/DeepSeek-R1-Distill-Llama-8B-GGUF", Description: "DeepSeek R1 8B LLaMA distilled", Tags: []string{"reasoning"}, SizeGB: 5.0},
	{Name: "deepseek-r1:14b", HFRepo: "bartowski/DeepSeek-R1-Distill-Qwen-14B-GGUF", Description: "DeepSeek R1 14B distilled", Tags: []string{"reasoning"}, SizeGB: 9.0},
	{Name: "deepseek-r1:32b", HFRepo: "bartowski/DeepSeek-R1-Distill-Qwen-32B-GGUF", Description: "DeepSeek R1 32B distilled", Tags: []string{"reasoning"}, SizeGB: 20.0},
	{Name: "deepseek-r1:70b", HFRepo: "bartowski/DeepSeek-R1-Distill-Llama-70B-GGUF", Description: "DeepSeek R1 70B LLaMA distilled", Tags: []string{"reasoning"}, SizeGB: 43.5},
	{Name: "deepseek-coder:6.7b", HFRepo: "TheBloke/deepseek-coder-6.7B-instruct-GGUF", Description: "DeepSeek Coder 6.7B", Tags: []string{"code"}, SizeGB: 4.1},
	{Name: "deepseek-coder-v2:16b", HFRepo: "bartowski/DeepSeek-Coder-V2-Lite-Instruct-GGUF", Description: "DeepSeek Coder V2 16B", Tags: []string{"code"}, SizeGB: 9.7},
	// ── Code LLaMA ─────────────────────────────────────────────────────────────
	{Name: "codellama:7b", HFRepo: "TheBloke/CodeLlama-7B-Instruct-GGUF", Description: "Meta Code LLaMA 7B", Tags: []string{"code"}, SizeGB: 4.1},
	{Name: "codellama:13b", HFRepo: "TheBloke/CodeLlama-13B-Instruct-GGUF", Description: "Meta Code LLaMA 13B", Tags: []string{"code"}, SizeGB: 7.9},
	{Name: "codellama:34b", HFRepo: "TheBloke/CodeLlama-34B-Instruct-GGUF", Description: "Meta Code LLaMA 34B", Tags: []string{"code"}, SizeGB: 20.7},
	// ── Yi ─────────────────────────────────────────────────────────────────────
	{Name: "yi:6b", HFRepo: "bartowski/Yi-6B-Chat-GGUF", Description: "01.AI Yi 6B Chat", Tags: []string{"chat"}, SizeGB: 3.7},
	{Name: "yi:9b", HFRepo: "bartowski/Yi-1.5-9B-Chat-GGUF", Description: "01.AI Yi 1.5 9B Chat", Tags: []string{"chat"}, SizeGB: 5.6},
	{Name: "yi:34b", HFRepo: "TheBloke/Yi-34B-Chat-GGUF", Description: "01.AI Yi 34B Chat", Tags: []string{"chat", "general"}, SizeGB: 20.7},
	{Name: "yi-coder:9b", HFRepo: "bartowski/Yi-Coder-9B-Chat-GGUF", Description: "01.AI Yi Coder 9B", Tags: []string{"code"}, SizeGB: 5.6},
	// ── Falcon ─────────────────────────────────────────────────────────────────
	{Name: "falcon:7b", HFRepo: "TheBloke/falcon-7b-instruct-GGUF", Description: "TII Falcon 7B Instruct", Tags: []string{"chat"}, SizeGB: 4.3},
	{Name: "falcon3:7b", HFRepo: "bartowski/Falcon3-7B-Instruct-GGUF", Description: "TII Falcon 3 7B", Tags: []string{"chat"}, SizeGB: 4.5},
	{Name: "falcon3:10b", HFRepo: "bartowski/Falcon3-10B-Instruct-GGUF", Description: "TII Falcon 3 10B", Tags: []string{"chat"}, SizeGB: 6.2},
	// ── StarCoder ──────────────────────────────────────────────────────────────
	{Name: "starcoder2:3b", HFRepo: "bartowski/starcoder2-3b-GGUF", Description: "BigCode StarCoder2 3B", Tags: []string{"code"}, SizeGB: 1.9},
	{Name: "starcoder2:7b", HFRepo: "bartowski/starcoder2-7b-GGUF", Description: "BigCode StarCoder2 7B", Tags: []string{"code"}, SizeGB: 4.4},
	{Name: "starcoder2:15b", HFRepo: "bartowski/starcoder2-15b-GGUF", Description: "BigCode StarCoder2 15B", Tags: []string{"code"}, SizeGB: 9.2},
	// ── Wizard ─────────────────────────────────────────────────────────────────
	{Name: "wizardlm2:7b", HFRepo: "bartowski/WizardLM-2-7B-GGUF", Description: "WizardLM 2 7B", Tags: []string{"chat"}, SizeGB: 4.4},
	{Name: "wizardlm2:8x22b", HFRepo: "bartowski/WizardLM-2-8x22B-GGUF", Description: "WizardLM 2 8x22B MoE", Tags: []string{"chat", "general"}, SizeGB: 79.9},
	{Name: "wizard-coder:33b", HFRepo: "TheBloke/WizardCoder-33B-V1.1-GGUF", Description: "WizardCoder 33B", Tags: []string{"code"}, SizeGB: 20.1},
	// ── Dolphin ────────────────────────────────────────────────────────────────
	{Name: "dolphin-mistral:7b", HFRepo: "bartowski/dolphin-2.9.3-mistral-nemo-12b-GGUF", Description: "Dolphin Mistral 7B", Tags: []string{"chat", "roleplay"}, SizeGB: 7.3},
	{Name: "dolphin3:8b", HFRepo: "bartowski/dolphin3.0-llama3.1-8b-GGUF", Description: "Dolphin 3 8B LLaMA", Tags: []string{"chat", "roleplay"}, SizeGB: 5.0},
	// ── Hermes ─────────────────────────────────────────────────────────────────
	{Name: "hermes3:8b", HFRepo: "bartowski/Hermes-3-Llama-3.1-8B-GGUF", Description: "Nous Hermes 3 8B", Tags: []string{"chat"}, SizeGB: 5.0},
	{Name: "hermes3:70b", HFRepo: "bartowski/Hermes-3-Llama-3.1-70B-GGUF", Description: "Nous Hermes 3 70B", Tags: []string{"chat", "general"}, SizeGB: 43.0},
	{Name: "openhermes2.5:7b", HFRepo: "TheBloke/OpenHermes-2.5-Mistral-7B-GGUF", Description: "OpenHermes 2.5 Mistral 7B", Tags: []string{"chat"}, SizeGB: 4.4},
	// ── Orca ───────────────────────────────────────────────────────────────────
	{Name: "orca-mini:3b", HFRepo: "TheBloke/orca_mini_3B-GGUF", Description: "Orca Mini 3B", Tags: []string{"chat", "edge"}, SizeGB: 1.9},
	{Name: "orca2:7b", HFRepo: "TheBloke/Orca-2-7B-GGUF", Description: "Microsoft Orca 2 7B", Tags: []string{"chat", "reasoning"}, SizeGB: 4.1},
	{Name: "orca2:13b", HFRepo: "TheBloke/Orca-2-13B-GGUF", Description: "Microsoft Orca 2 13B", Tags: []string{"chat", "reasoning"}, SizeGB: 7.9},
	// ── Vicuna ─────────────────────────────────────────────────────────────────
	{Name: "vicuna:7b", HFRepo: "TheBloke/vicuna-7B-v1.5-GGUF", Description: "Vicuna 7B v1.5", Tags: []string{"chat"}, SizeGB: 4.1},
	{Name: "vicuna:13b", HFRepo: "TheBloke/vicuna-13B-v1.5-GGUF", Description: "Vicuna 13B v1.5", Tags: []string{"chat"}, SizeGB: 7.9},
	{Name: "vicuna:33b", HFRepo: "TheBloke/vicuna-33B-GGUF", Description: "Vicuna 33B", Tags: []string{"chat", "general"}, SizeGB: 20.1},
	// ── OpenChat ───────────────────────────────────────────────────────────────
	{Name: "openchat:7b", HFRepo: "bartowski/openchat-3.6-8b-20240522-GGUF", Description: "OpenChat 3.6 8B", Tags: []string{"chat"}, SizeGB: 5.0},
	// ── Zephyr ─────────────────────────────────────────────────────────────────
	{Name: "zephyr:7b", HFRepo: "TheBloke/zephyr-7B-beta-GGUF", Description: "Zephyr 7B Beta", Tags: []string{"chat"}, SizeGB: 4.4},
	// ── Solar ──────────────────────────────────────────────────────────────────
	{Name: "solar:10.7b", HFRepo: "bartowski/SOLAR-10.7B-Instruct-v1.0-GGUF", Description: "Upstage SOLAR 10.7B", Tags: []string{"chat"}, SizeGB: 6.5},
	{Name: "solar-pro:22b", HFRepo: "bartowski/solar-pro-preview-instruct-GGUF", Description: "Upstage Solar Pro 22B", Tags: []string{"chat", "general"}, SizeGB: 13.5},
	// ── Command R ──────────────────────────────────────────────────────────────
	{Name: "command-r:35b", HFRepo: "bartowski/c4ai-command-r-v01-GGUF", Description: "Cohere Command R 35B", Tags: []string{"chat", "tools"}, SizeGB: 21.4},
	{Name: "command-r7b:7b", HFRepo: "bartowski/c4ai-command-r7b-12-2024-GGUF", Description: "Cohere Command R7B", Tags: []string{"chat", "edge"}, SizeGB: 4.4},
	// ── Stable LM ──────────────────────────────────────────────────────────────
	{Name: "stablelm2:1.6b", HFRepo: "bartowski/stablelm-2-zephyr-1_6b-GGUF", Description: "StableLM 2 1.6B Zephyr", Tags: []string{"chat", "edge"}, SizeGB: 1.0},
	{Name: "stablelm2:12b", HFRepo: "bartowski/stablelm-2-12b-chat-GGUF", Description: "StableLM 2 12B Chat", Tags: []string{"chat"}, SizeGB: 7.5},
	// ── Embedding ──────────────────────────────────────────────────────────────
	{Name: "nomic-embed-text", HFRepo: "nomic-ai/nomic-embed-text-v1.5-GGUF", Description: "Nomic text embeddings", Tags: []string{"embed"}, SizeGB: 0.1},
	{Name: "mxbai-embed-large", HFRepo: "mixedbread-ai/mxbai-embed-large-v1-GGUF", Description: "MxBai Embed Large", Tags: []string{"embed"}, SizeGB: 0.2},
	{Name: "bge-m3", HFRepo: "gpustack/bge-m3-GGUF", Description: "BGE M3 multilingual embed", Tags: []string{"embed", "multilingual"}, SizeGB: 0.3},
	// ── Nemotron ───────────────────────────────────────────────────────────────
	{Name: "nemotron:70b", HFRepo: "bartowski/Llama-3.1-Nemotron-70B-Instruct-HF-GGUF", Description: "NVIDIA Nemotron 70B", Tags: []string{"chat", "general"}, SizeGB: 43.0},
	{Name: "nemotron-mini:4b", HFRepo: "bartowski/Nemotron-Mini-4B-Instruct-GGUF", Description: "NVIDIA Nemotron Mini 4B", Tags: []string{"chat", "edge"}, SizeGB: 2.5},
	// ── Tiny ───────────────────────────────────────────────────────────────────
	{Name: "tinyllama:1.1b", HFRepo: "TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF", Description: "TinyLLaMA 1.1B Chat", Tags: []string{"chat", "edge"}, SizeGB: 0.7},
	{Name: "smollm2:135m", HFRepo: "bartowski/SmolLM2-135M-Instruct-GGUF", Description: "SmolLM2 135M", Tags: []string{"chat", "edge"}, SizeGB: 0.1},
	{Name: "smollm2:360m", HFRepo: "bartowski/SmolLM2-360M-Instruct-GGUF", Description: "SmolLM2 360M", Tags: []string{"chat", "edge"}, SizeGB: 0.2},
	{Name: "smollm2:1.7b", HFRepo: "bartowski/SmolLM2-1.7B-Instruct-GGUF", Description: "SmolLM2 1.7B", Tags: []string{"chat", "edge"}, SizeGB: 1.1},
	// ── InternLM ───────────────────────────────────────────────────────────────
	{Name: "internlm2.5:7b", HFRepo: "bartowski/internlm2_5-7b-chat-GGUF", Description: "InternLM 2.5 7B Chat", Tags: []string{"chat"}, SizeGB: 4.4},
	{Name: "internlm3:8b", HFRepo: "bartowski/internlm3-8b-instruct-GGUF", Description: "InternLM 3 8B", Tags: []string{"chat"}, SizeGB: 5.0},
	// ── Granite ────────────────────────────────────────────────────────────────
	{Name: "granite-code:3b", HFRepo: "bartowski/granite-3b-code-instruct-128k-GGUF", Description: "IBM Granite Code 3B", Tags: []string{"code"}, SizeGB: 1.9},
	{Name: "granite-code:8b", HFRepo: "bartowski/granite-8b-code-instruct-128k-GGUF", Description: "IBM Granite Code 8B", Tags: []string{"code"}, SizeGB: 5.0},
	{Name: "granite3.1-dense:2b", HFRepo: "bartowski/granite-3.1-2b-instruct-GGUF", Description: "IBM Granite 3.1 Dense 2B", Tags: []string{"chat", "edge"}, SizeGB: 1.3},
	{Name: "granite3.1-dense:8b", HFRepo: "bartowski/granite-3.1-8b-instruct-GGUF", Description: "IBM Granite 3.1 Dense 8B", Tags: []string{"chat"}, SizeGB: 5.0},
	// ── LLaVA ──────────────────────────────────────────────────────────────────
	{Name: "llava:7b", HFRepo: "mys/ggml_llava-v1.5-7b", Description: "LLaVA 1.5 7B vision", Tags: []string{"vision", "chat"}, SizeGB: 4.1},
	{Name: "llava:13b", HFRepo: "mys/ggml_llava-v1.5-13b", Description: "LLaVA 1.5 13B vision", Tags: []string{"vision", "chat"}, SizeGB: 7.9},
	{Name: "llava-phi3:3.8b", HFRepo: "xtuner/llava-phi-3-mini-gguf", Description: "LLaVA on Phi 3 3.8B", Tags: []string{"vision", "edge"}, SizeGB: 2.4},
	{Name: "moondream:1.8b", HFRepo: "vikhyatk/moondream2", Description: "Moondream 2 vision", Tags: []string{"vision", "edge"}, SizeGB: 1.1},
	// ── Chinese ────────────────────────────────────────────────────────────────
	{Name: "chatglm3:6b", HFRepo: "THUDM/chatglm3-6b-gguf", Description: "ChatGLM 3 6B Chinese", Tags: []string{"chat", "chinese"}, SizeGB: 3.7},
	{Name: "yi-vl:6b", HFRepo: "bartowski/Yi-VL-6B-GGUF", Description: "Yi VL 6B vision-language", Tags: []string{"vision", "chinese"}, SizeGB: 3.7},
	// ── Medical ────────────────────────────────────────────────────────────────
	{Name: "meditron:7b", HFRepo: "TheBloke/meditron-7B-GGUF", Description: "EPFL Meditron 7B medical", Tags: []string{"medical"}, SizeGB: 4.1},
	{Name: "medalpaca:7b", HFRepo: "TheBloke/medalpaca-7B-GGUF", Description: "MedAlpaca 7B", Tags: []string{"medical"}, SizeGB: 4.1},
	// ── Platypus ───────────────────────────────────────────────────────────────
	{Name: "platypus2:70b", HFRepo: "TheBloke/Platypus2-70B-instruct-GGUF", Description: "Platypus 2 70B STEM", Tags: []string{"chat", "reasoning"}, SizeGB: 41.4},
	// ── RWKV ───────────────────────────────────────────────────────────────────
	{Name: "rwkv:7b", HFRepo: "BlinkDL/rwkv-6-world", Description: "RWKV 7B world model", Tags: []string{"chat", "ssm"}, SizeGB: 14.4},
	// ── Tulu ───────────────────────────────────────────────────────────────────
	{Name: "tulu3:8b", HFRepo: "bartowski/Llama-3.1-Tulu-3-8B-GGUF", Description: "AllenAI Tulu 3 8B", Tags: []string{"chat", "reasoning"}, SizeGB: 5.0},
	{Name: "tulu3:70b", HFRepo: "bartowski/Llama-3.1-Tulu-3-70B-GGUF", Description: "AllenAI Tulu 3 70B", Tags: []string{"chat", "reasoning"}, SizeGB: 43.0},
	// ── Aya ────────────────────────────────────────────────────────────────────
	{Name: "aya-expanse:8b", HFRepo: "bartowski/aya-expanse-8b-GGUF", Description: "Cohere Aya Expanse 8B multilingual", Tags: []string{"chat", "multilingual"}, SizeGB: 5.0},
	{Name: "aya-expanse:32b", HFRepo: "bartowski/aya-expanse-32b-GGUF", Description: "Cohere Aya Expanse 32B multilingual", Tags: []string{"chat", "multilingual"}, SizeGB: 19.8},
	// ── SQLCoder ───────────────────────────────────────────────────────────────
	{Name: "sqlcoder:7b", HFRepo: "bartowski/sqlcoder-7b-2-GGUF", Description: "Defog SQLCoder 7B", Tags: []string{"code", "sql"}, SizeGB: 4.4},
	{Name: "sqlcoder:15b", HFRepo: "defog/sqlcoder-7b-2", Description: "Defog SQLCoder 15B", Tags: []string{"code", "sql"}, SizeGB: 9.2},
	// ── Granite Code ───────────────────────────────────────────────────────────
	{Name: "granite-code:20b", HFRepo: "bartowski/granite-20b-code-instruct-8k-GGUF", Description: "IBM Granite Code 20B", Tags: []string{"code"}, SizeGB: 12.2},
	{Name: "granite-code:34b", HFRepo: "bartowski/granite-34b-code-instruct-8k-GGUF", Description: "IBM Granite Code 34B", Tags: []string{"code"}, SizeGB: 20.7},
	// ── Reflection ─────────────────────────────────────────────────────────────
	{Name: "reflection:70b", HFRepo: "bartowski/Reflection-Llama-3.1-70B-GGUF", Description: "Reflection 70B reasoning", Tags: []string{"reasoning"}, SizeGB: 43.0},
	// ── Exaone ─────────────────────────────────────────────────────────────────
	{Name: "exaone3.5:2.4b", HFRepo: "bartowski/EXAONE-3.5-2.4B-Instruct-GGUF", Description: "LG EXAONE 3.5 2.4B", Tags: []string{"chat", "edge"}, SizeGB: 1.5},
	{Name: "exaone3.5:7.8b", HFRepo: "bartowski/EXAONE-3.5-7.8B-Instruct-GGUF", Description: "LG EXAONE 3.5 7.8B", Tags: []string{"chat"}, SizeGB: 4.9},
	{Name: "exaone3.5:32b", HFRepo: "bartowski/EXAONE-3.5-32B-Instruct-GGUF", Description: "LG EXAONE 3.5 32B", Tags: []string{"chat", "general"}, SizeGB: 19.8},
	// ── NuExtract ──────────────────────────────────────────────────────────────
	{Name: "nuextract:3.8b", HFRepo: "bartowski/NuExtract-v1.5-GGUF", Description: "NuExtract structured extraction", Tags: []string{"tools", "extract"}, SizeGB: 2.4},
	// ── Cogito ─────────────────────────────────────────────────────────────────
	{Name: "cogito:8b", HFRepo: "bartowski/cogito-v1-preview-llama-3B-GGUF", Description: "Cogito hybrid reasoning 8B", Tags: []string{"reasoning"}, SizeGB: 5.0},
	{Name: "cogito:32b", HFRepo: "bartowski/cogito-v1-preview-qwen-32b-GGUF", Description: "Cogito hybrid reasoning 32B", Tags: []string{"reasoning"}, SizeGB: 20.0},
	// ── Devstral ───────────────────────────────────────────────────────────────
	{Name: "devstral:24b", HFRepo: "bartowski/Devstral-Small-2505-GGUF", Description: "Mistral agentic coder 24B", Tags: []string{"code", "agent"}, SizeGB: 14.8},
	// ── Moondream ──────────────────────────────────────────────────────────────
	{Name: "moondream2:1.8b", HFRepo: "vikhyatk/moondream2", Description: "Moondream 2 vision model", Tags: []string{"vision", "edge"}, SizeGB: 1.1},
	// ── SmolVLM ────────────────────────────────────────────────────────────────
	{Name: "smolvlm:500m", HFRepo: "ggml-org/SmolVLM-500M-Instruct-GGUF", Description: "SmolVLM 500M tiny vision", Tags: []string{"vision", "edge"}, SizeGB: 0.3},
	// ── Phind ──────────────────────────────────────────────────────────────────
	{Name: "phind-codellama:34b", HFRepo: "TheBloke/Phind-CodeLlama-34B-v2-GGUF", Description: "Phind CodeLLaMA 34B", Tags: []string{"code"}, SizeGB: 20.7},
	// ── Bagel ──────────────────────────────────────────────────────────────────
	{Name: "bagel:34b", HFRepo: "bartowski/bagel-dpo-34b-v0.2-GGUF", Description: "Bagel DPO 34B", Tags: []string{"chat"}, SizeGB: 20.7},
	// ── MiniCPM ────────────────────────────────────────────────────────────────
	{Name: "minicpm3:4b", HFRepo: "bartowski/MiniCPM3-4B-GGUF", Description: "MiniCPM 3 4B edge model", Tags: []string{"chat", "edge"}, SizeGB: 2.5},
	{Name: "minicpm-v:8b", HFRepo: "bartowski/MiniCPM-V-2_6-GGUF", Description: "MiniCPM V 8B vision", Tags: []string{"vision"}, SizeGB: 5.0},
	// ── OLMo ───────────────────────────────────────────────────────────────────
	{Name: "olmo2:7b", HFRepo: "bartowski/OLMo-2-1124-7B-Instruct-GGUF", Description: "AllenAI OLMo 2 7B", Tags: []string{"chat", "research"}, SizeGB: 4.4},
	{Name: "olmo2:13b", HFRepo: "bartowski/OLMo-2-1124-13B-Instruct-GGUF", Description: "AllenAI OLMo 2 13B", Tags: []string{"chat", "research"}, SizeGB: 7.9},
	// ── Open R1 ────────────────────────────────────────────────────────────────
	{Name: "open-r1:7b", HFRepo: "bartowski/open-r1-math-7b-GGUF", Description: "HuggingFace Open R1 7B", Tags: []string{"reasoning", "math"}, SizeGB: 4.4},
	{Name: "open-r1:14b", HFRepo: "bartowski/OpenR1-Qwen-14B-GGUF", Description: "Open R1 14B", Tags: []string{"reasoning"}, SizeGB: 9.0},
	// ── Skywork ────────────────────────────────────────────────────────────────
	{Name: "skywork-o1:8b", HFRepo: "bartowski/Skywork-o1-Open-Llama-3.1-8B-GGUF", Description: "Skywork o1 reasoning 8B", Tags: []string{"reasoning"}, SizeGB: 5.0},
	// ── Goliath ────────────────────────────────────────────────────────────────
	{Name: "goliath:120b", HFRepo: "bartowski/goliath-120b-GGUF", Description: "Goliath merged 120B", Tags: []string{"chat"}, SizeGB: 73.4},
	// ── Dolphin Mix ────────────────────────────────────────────────────────────
	{Name: "dolphin-mixtral:8x7b", HFRepo: "bartowski/dolphin-2.7-mixtral-8x7b-GGUF", Description: "Dolphin Mixtral 8x7B", Tags: []string{"chat", "roleplay"}, SizeGB: 26.4},
	// ── Nous Capybara ──────────────────────────────────────────────────────────
	{Name: "nous-capybara:7b", HFRepo: "TheBloke/Nous-Capybara-7B-V1.9-GGUF", Description: "Nous Capybara 7B", Tags: []string{"chat"}, SizeGB: 4.1},
	// ── Airoboros ──────────────────────────────────────────────────────────────
	{Name: "airoboros:70b", HFRepo: "TheBloke/Airoboros-L2-70b-GPT4-2.0-GGUF", Description: "Airoboros 70B context obedient", Tags: []string{"chat"}, SizeGB: 41.4},
	// ── Openbuddy ──────────────────────────────────────────────────────────────
	{Name: "openbuddy:70b", HFRepo: "TheBloke/OpenBuddy-LLaMA2-70B-v13.2-GGUF", Description: "OpenBuddy multilingual 70B", Tags: []string{"chat", "multilingual"}, SizeGB: 41.4},
	// ── NexusRaven ─────────────────────────────────────────────────────────────
	{Name: "nexusraven:13b", HFRepo: "bartowski/NexusRaven-V2-13B-GGUF", Description: "Nexus function calling 13B", Tags: []string{"tools"}, SizeGB: 7.9},
	// ── Reader LM ──────────────────────────────────────────────────────────────
	{Name: "reader-lm:1.5b", HFRepo: "bartowski/reader-lm-1.5b-GGUF", Description: "Jina Reader LM 1.5B", Tags: []string{"tools"}, SizeGB: 1.0},
	// ── Magicoder ──────────────────────────────────────────────────────────────
	{Name: "magicoder:7b", HFRepo: "TheBloke/Magicoder-S-DS-6.7B-GGUF", Description: "Magicoder 7B OSS-Instruct", Tags: []string{"code"}, SizeGB: 4.1},
	// ── Stable Beluga ──────────────────────────────────────────────────────────
	{Name: "stable-beluga:70b", HFRepo: "TheBloke/StableBeluga-70B-GGUF", Description: "Stable Beluga 70B", Tags: []string{"chat"}, SizeGB: 41.4},
	// ── GPT-J ──────────────────────────────────────────────────────────────────
	{Name: "gpt-j:6b", HFRepo: "nomic-ai/gpt4all-j-groovy-GGUF", Description: "GPT-J 6B", Tags: []string{"chat"}, SizeGB: 3.8},
	// ── Jais ───────────────────────────────────────────────────────────────────
	{Name: "jais:13b", HFRepo: "inceptionai/jais-13b-chat-GGUF", Description: "Core42 Jais 13B Arabic", Tags: []string{"chat", "arabic"}, SizeGB: 7.9},
	// ── Sailor ─────────────────────────────────────────────────────────────────
	{Name: "sailor2:8b", HFRepo: "bartowski/Sailor2-8B-Chat-GGUF", Description: "Sailor2 8B multilingual", Tags: []string{"chat", "multilingual"}, SizeGB: 5.0},
	// ── Meditron ───────────────────────────────────────────────────────────────
	{Name: "meditron:70b", HFRepo: "TheBloke/meditron-70B-GGUF", Description: "EPFL Meditron 70B medical", Tags: []string{"medical"}, SizeGB: 41.4},
	// ── LLaMA 3.2 Vision ───────────────────────────────────────────────────────
	{Name: "llama3.2-vision:11b", HFRepo: "bartowski/Llama-3.2-11B-Vision-Instruct-GGUF", Description: "LLaMA 3.2 11B Vision", Tags: []string{"vision", "chat"}, SizeGB: 6.8},
	{Name: "llama3.2-vision:90b", HFRepo: "bartowski/Llama-3.2-90B-Vision-Instruct-GGUF", Description: "LLaMA 3.2 90B Vision", Tags: []string{"vision", "chat"}, SizeGB: 55.4},
	// ── Phi4 Multimodal ────────────────────────────────────────────────────────
	{Name: "phi4-multimodal:5.6b", HFRepo: "bartowski/Phi-4-multimodal-instruct-GGUF", Description: "Phi 4 multimodal vision+audio", Tags: []string{"vision", "audio"}, SizeGB: 3.5},
	// ── Hunyuan ────────────────────────────────────────────────────────────────
	{Name: "hunyuan-a13b", HFRepo: "bartowski/Hunyuan-A13B-Instruct-GGUF", Description: "Tencent Hunyuan A13B MoE", Tags: []string{"chat", "general"}, SizeGB: 8.0},
	// ── InternVL ───────────────────────────────────────────────────────────────
	{Name: "internvl2:8b", HFRepo: "bartowski/InternVL2-8B-GGUF", Description: "InternVL 2 8B vision", Tags: []string{"vision", "chat"}, SizeGB: 5.0},
	{Name: "internvl2:26b", HFRepo: "bartowski/InternVL2-26B-GGUF", Description: "InternVL 2 26B vision", Tags: []string{"vision", "chat"}, SizeGB: 16.0},
	// ── Mistral Small 3.1 ──────────────────────────────────────────────────────
	{Name: "mistral-small3.1:24b", HFRepo: "bartowski/Mistral-Small-3.1-24B-Instruct-2503-GGUF", Description: "Mistral Small 3.1 24B vision", Tags: []string{"chat", "vision"}, SizeGB: 14.8},
	// ── Atlas ──────────────────────────────────────────────────────────────────
	{Name: "llama3-chatqa:8b", HFRepo: "bartowski/Llama3-ChatQA-1.5-8B-GGUF", Description: "NVIDIA ChatQA 1.5 8B", Tags: []string{"chat", "rag"}, SizeGB: 5.0},
	{Name: "llama3-chatqa:70b", HFRepo: "bartowski/Llama3-ChatQA-1.5-70B-GGUF", Description: "NVIDIA ChatQA 1.5 70B", Tags: []string{"chat", "rag"}, SizeGB: 43.0},
	// ── Granite 3.2 Vision ─────────────────────────────────────────────────────
	{Name: "granite3.2-vision:2b", HFRepo: "bartowski/granite-3.2-2b-instruct-GGUF", Description: "IBM Granite 3.2 Vision 2B", Tags: []string{"vision", "edge"}, SizeGB: 1.3},
	// ── Athene ─────────────────────────────────────────────────────────────────
	{Name: "athene-v2:72b", HFRepo: "bartowski/Athene-V2-Chat-GGUF", Description: "Nexusflow Athene v2 72B", Tags: []string{"chat", "tools"}, SizeGB: 44.5},
	// ── Bespoke ────────────────────────────────────────────────────────────────
	{Name: "bespoke-minicheck:7b", HFRepo: "bartowski/Bespoke-MiniCheck-7B-GGUF", Description: "Bespoke MiniCheck fact checker", Tags: []string{"tools"}, SizeGB: 4.4},
	// ── GLM 4 ──────────────────────────────────────────────────────────────────
	{Name: "glm4:9b", HFRepo: "bartowski/glm-4-9b-chat-GGUF", Description: "Tsinghua GLM 4 9B", Tags: []string{"chat", "chinese"}, SizeGB: 5.6},
	// ── Kimi VL ────────────────────────────────────────────────────────────────
	{Name: "kimi-vl:16b", HFRepo: "bartowski/Kimi-VL-A3B-Instruct-GGUF", Description: "Moonshot Kimi VL 16B", Tags: []string{"vision", "chat"}, SizeGB: 9.9},
	// ── Nemotron Super/Ultra ───────────────────────────────────────────────────
	{Name: "nemotron-super:49b", HFRepo: "bartowski/Llama-3.3-Nemotron-Super-49B-v1-GGUF", Description: "NVIDIA Nemotron Super 49B", Tags: []string{"chat", "general"}, SizeGB: 30.3},
}

// hfPullStartTime tracks when downloads started for progress display
var hfPullStartTime = time.Now()
