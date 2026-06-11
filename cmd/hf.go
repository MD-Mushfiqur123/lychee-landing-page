package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// NewHFCmd creates the "lychee hf" subcommand group
func NewHFCmd() *cobra.Command {
	hfCmd := &cobra.Command{
		Use:   "hf",
		Short: "HuggingFace model management (275+ models, no token required)",
		Long: `Pull and run AI models directly from HuggingFace Hub.
No account or token required for public models.
All models are downloaded in GGUF format for local inference.

Examples:
  lychee hf pull microsoft/Phi-3-mini-4k-instruct-gguf
  lychee hf pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --quant q5_k_m
  lychee hf pull bartowski/Mixtral-8x7B-Instruct-v0.1-GGUF --list
  lychee hf search code
  lychee hf list`,
	}

	var quantFlag string
	var listFlag bool

	pullCmd := &cobra.Command{
		Use:   "pull <org/repo>",
		Short: "Pull a model from HuggingFace (resumable, verified, no token needed)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return HFPullHandler(cmd, args, quantFlag, listFlag)
		},
	}
	pullCmd.Flags().StringVar(&quantFlag, "quant", "", "quantization to prefer (e.g. q4_k_m, q5_k_m, q8_0)")
	pullCmd.Flags().BoolVar(&listFlag, "list", false, "list all available quantizations then exit")

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search the built-in Lychee model catalog",
		Args:  cobra.MaximumNArgs(1),
		RunE:  HFSearchHandler,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all models in the Lychee catalog",
		Args:  cobra.NoArgs,
		RunE:  HFListHandler,
	}

	hfCmd.AddCommand(pullCmd, searchCmd, listCmd)
	return hfCmd
}

// HFPullHandler handles "lychee hf pull <hf-model-ref>"
func HFPullHandler(cmd *cobra.Command, args []string, quant string, listOnly bool) error {
	modelRef := args[0]
	if !strings.Contains(modelRef, "/") {
		return fmt.Errorf("invalid HuggingFace model reference %q\nExpected format: org/repo or hf://org/repo", modelRef)
	}

	modelDir := filepath.Join(os.Getenv("USERPROFILE"), ".lychee", "models")
	if h := os.Getenv("HOME"); h != "" && modelDir == filepath.Join("", ".lychee", "models") {
		modelDir = filepath.Join(h, ".lychee", "models")
	}
	if lycheeModels := os.Getenv("LYCHEE_MODELS"); lycheeModels != "" {
		modelDir = lycheeModels
	}

	return hfPullModel(cmd.Context(), modelRef, modelDir, quant, listOnly)
}

func hfPullModel(ctx context.Context, modelRef, modelDir, quant string, listOnly bool) error {
	ref := strings.TrimPrefix(modelRef, "hf://")
	ref = strings.TrimPrefix(ref, "hf:")
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected org/repo format, got: %s", modelRef)
	}
	org, repo := parts[0], parts[1]

	client := &hfHTTPClient{
		apiClient:      &http.Client{Timeout: 30 * time.Second},
		downloadClient: &http.Client{},
	}

	fmt.Printf("  Fetching model info: %s/%s\n", org, repo)
	files, err := client.listGGUFFiles(ctx, org, repo)
	if err != nil {
		return fmt.Errorf("fetching model info: %w\nMake sure the model exists and is public, or set LYCHEE_HF_TOKEN for private models", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no GGUF files found in %s/%s\nTip: use 'lychee hf search %s' to find compatible repos", org, repo, repo)
	}

	// Group files by quantization
	quants := groupByQuant(files)

	if listOnly {
		printQuantTable(org, repo, quants)
		return nil
	}

	// Select files to download
	var selected []hfFile
	if quant != "" {
		q := strings.ToLower(quant)
		selected = quants[q]
		if len(selected) == 0 {
			fmt.Printf("  Quantization %q not found. Available:\n\n", quant)
			printQuantTable(org, repo, quants)
			return fmt.Errorf("quantization %q not available", quant)
		}
	} else {
		selected = selectBestQuant(quants)
	}

	// Sort shards by name so they join in order
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Name < selected[j].Name
	})

	destDir := filepath.Join(modelDir, "hf-models", org, repo)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	fmt.Printf("\n  Model : %s/%s\n", org, repo)
	if len(selected) == 1 {
		fmt.Printf("  File  : %s\n", selected[0].Name)
	} else {
		fmt.Printf("  Shards: %d files\n", len(selected))
		for _, f := range selected {
			fmt.Printf("          %s\n", f.Name)
		}
	}
	totalBytes := int64(0)
	for _, f := range selected {
		totalBytes += f.Size
	}
	if totalBytes > 0 {
		fmt.Printf("  Size  : %s\n", formatBytes(totalBytes))
	}
	fmt.Println()

	// Download — parallel for multi-shard, sequential for single
	if len(selected) > 1 {
		if err := downloadShards(ctx, client, org, repo, selected, destDir); err != nil {
			return err
		}
	} else {
		f := selected[0]
		dest := filepath.Join(destDir, f.Name)
		if err := downloadWithResume(ctx, client, org, repo, "main", f, dest); err != nil {
			return err
		}
	}

	// Write Modelfile
	var fromPaths []string
	for _, f := range selected {
		fromPaths = append(fromPaths, filepath.Join(destDir, f.Name))
	}
	modelfile := "FROM " + fromPaths[0] + "\n"
	mfPath := filepath.Join(destDir, "Modelfile")
	_ = os.WriteFile(mfPath, []byte(modelfile), 0o644)

	fmt.Printf("\n  Done. To run:\n")
	fmt.Printf("    lychee create %s-%s -f %s\n", repo, selected[0].quantKey, mfPath)
	fmt.Printf("    lychee run %s-%s\n", repo, selected[0].quantKey)
	return nil
}

// downloadShards downloads multiple GGUF shards in parallel (up to 3 at once)
func downloadShards(ctx context.Context, client *hfHTTPClient, org, repo string, shards []hfFile, destDir string) error {
	type result struct {
		name string
		err  error
	}

	sem := make(chan struct{}, 3) // max 3 parallel
	results := make(chan result, len(shards))
	var wg sync.WaitGroup

	for _, f := range shards {
		wg.Add(1)
		go func(file hfFile) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			dest := filepath.Join(destDir, file.Name)
			err := downloadWithResume(ctx, client, org, repo, "main", file, dest)
			results <- result{file.Name, err}
		}(f)
	}

	wg.Wait()
	close(results)

	for r := range results {
		if r.err != nil {
			return fmt.Errorf("shard %s: %w", r.name, r.err)
		}
	}
	return nil
}

// downloadWithResume downloads a file with HTTP Range resume support and SHA256 verification
func downloadWithResume(ctx context.Context, client *hfHTTPClient, org, repo, revision string, f hfFile, destPath string) error {
	tmpPath := destPath + ".part"

	// Check existing partial download
	var startByte int64
	if info, err := os.Stat(tmpPath); err == nil {
		startByte = info.Size()
		if f.Size > 0 && startByte >= f.Size {
			// Already complete — just rename
			return os.Rename(tmpPath, destPath)
		}
		if startByte > 0 {
			fmt.Printf("  Resuming %s from %s\n", f.Name, formatBytes(startByte))
		}
	}

	// Skip if already fully downloaded and verified
	if _, err := os.Stat(destPath); err == nil && f.SHA256 != "" {
		if verifySHA256(destPath, f.SHA256) {
			fmt.Printf("  ✓ %s already downloaded and verified\n", f.Name)
			return nil
		}
		// Hash mismatch — redownload
		_ = os.Remove(destPath)
		startByte = 0
	}

	url := fmt.Sprintf("https://huggingface.co/%s/%s/resolve/%s/%s", org, repo, revision, f.Name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Lychee/1.0")
	if tok := os.Getenv("LYCHEE_HF_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	resp, err := client.doDownload(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, f.Name)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if startByte > 0 && resp.StatusCode == 206 {
		flags |= os.O_APPEND
	} else {
		startByte = 0 // server didn't honor range, start fresh
	}

	out, err := os.OpenFile(tmpPath, flags, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	total := f.Size
	if total == 0 {
		total = resp.ContentLength
	}

	bar := newProgressBar(f.Name, startByte, total)
	hasher := sha256.New()

	// If resuming, we can't hash the already-downloaded part from scratch
	// so we only verify if starting fresh
	canVerify := startByte == 0 && f.SHA256 != ""

	buf := make([]byte, 1<<17) // 128KB chunks
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
			if canVerify {
				hasher.Write(buf[:n])
			}
			bar.add(int64(n))
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	bar.finish()
	out.Close()

	// SHA256 verification
	if canVerify {
		got := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(got, f.SHA256) {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("SHA256 mismatch for %s\n  expected: %s\n  got:      %s", f.Name, f.SHA256, got)
		}
		fmt.Printf("  ✓ SHA256 verified: %s\n", f.Name)
	}

	return os.Rename(tmpPath, destPath)
}

// ── Progress bar ──────────────────────────────────────────────────────────────

type progressBar struct {
	name      string
	total     int64
	done      int64
	startTime time.Time
	mu        sync.Mutex
}

func newProgressBar(name string, already, total int64) *progressBar {
	p := &progressBar{name: name, total: total, done: already, startTime: time.Now()}
	p.render()
	return p
}

func (p *progressBar) add(n int64) {
	p.mu.Lock()
	p.done += n
	p.mu.Unlock()
	p.render()
}

func (p *progressBar) render() {
	p.mu.Lock()
	done := p.done
	total := p.total
	p.mu.Unlock()

	elapsed := time.Since(p.startTime).Seconds()
	speed := float64(0)
	if elapsed > 0 {
		speed = float64(done) / elapsed
	}

	var pct float64
	var bar string
	if total > 0 {
		pct = float64(done) / float64(total) * 100
		filled := int(pct / 5) // 20-char bar
		bar = "[" + strings.Repeat("█", filled) + strings.Repeat("░", 20-filled) + "]"
	} else {
		bar = "[" + strings.Repeat("░", 20) + "]"
	}

	eta := ""
	if speed > 0 && total > done {
		secs := int((float64(total-done)) / speed)
		if secs < 60 {
			eta = fmt.Sprintf(" ETA %ds", secs)
		} else {
			eta = fmt.Sprintf(" ETA %dm%ds", secs/60, secs%60)
		}
	}

	name := p.name
	if len(name) > 30 {
		name = "..." + name[len(name)-27:]
	}

	if total > 0 {
		fmt.Printf("\r  %-30s %s %5.1f%%  %s/s%s  %s / %s   ",
			name, bar, pct,
			formatBytes(int64(speed)),
			eta,
			formatBytes(done),
			formatBytes(total),
		)
	} else {
		fmt.Printf("\r  %-30s  %s  %s/s   ",
			name,
			formatBytes(done),
			formatBytes(int64(speed)),
		)
	}
}

func (p *progressBar) finish() {
	p.mu.Lock()
	p.done = p.total
	p.mu.Unlock()
	p.render()
	fmt.Println()
}

// ── HF file types & grouping ──────────────────────────────────────────────────

type hfFile struct {
	Name     string
	Size     int64
	SHA256   string
	quantKey string // e.g. "q4_k_m"
}

func groupByQuant(files []hfFile) map[string][]hfFile {
	groups := make(map[string][]hfFile)
	for _, f := range files {
		groups[f.quantKey] = append(groups[f.quantKey], f)
	}
	return groups
}

func printQuantTable(org, repo string, quants map[string][]hfFile) {
	fmt.Printf("  Available quantizations for %s/%s:\n\n", org, repo)
	fmt.Printf("  %-16s %-8s %s\n", "QUANT", "SIZE", "FILES")
	fmt.Println("  " + strings.Repeat("-", 60))

	// Sort keys
	keys := make([]string, 0, len(quants))
	for k := range quants {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	priority := []string{"q4_k_m", "q4_k_s", "q5_k_m", "q4_0", "q5_0", "q6_k", "q8_0", "f16", "fp16"}
	printed := map[string]bool{}

	// Print in priority order first
	for _, p := range priority {
		if fs, ok := quants[p]; ok {
			total := int64(0)
			for _, f := range fs { total += f.Size }
			marker := ""
			if p == "q4_k_m" { marker = " ← recommended" }
			fmt.Printf("  %-16s %-8s %d file(s)%s\n", p, formatBytes(total), len(fs), marker)
			printed[p] = true
		}
	}
	// Print remaining
	for _, k := range keys {
		if !printed[k] {
			fs := quants[k]
			total := int64(0)
			for _, f := range fs { total += f.Size }
			fmt.Printf("  %-16s %-8s %d file(s)\n", k, formatBytes(total), len(fs))
		}
	}
	fmt.Printf("\n  Pull with: lychee hf pull %s/%s --quant <name>\n\n", org, repo)
}

func selectBestQuant(quants map[string][]hfFile) []hfFile {
	priority := []string{"q4_k_m", "q4_k_s", "q5_k_m", "q4_0", "q5_0", "q6_k", "q8_0", "f16", "fp16"}
	for _, p := range priority {
		if fs, ok := quants[p]; ok {
			return fs
		}
	}
	// Fallback: first key alphabetically
	keys := make([]string, 0, len(quants))
	for k := range quants { keys = append(keys, k) }
	sort.Strings(keys)
	if len(keys) > 0 {
		return quants[keys[0]]
	}
	return nil
}

// ── HTTP client ───────────────────────────────────────────────────────────────

type hfHTTPClient struct {
	apiClient      *http.Client
	downloadClient *http.Client
}

func (h *hfHTTPClient) doAPI(req *http.Request) (*http.Response, error) {
	resp, err := h.apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		resp.Body.Close()
		return nil, fmt.Errorf("private model — set LYCHEE_HF_TOKEN env var")
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		return nil, fmt.Errorf("model not found on HuggingFace")
	}
	return resp, nil
}

func (h *hfHTTPClient) doDownload(req *http.Request) (*http.Response, error) {
	resp, err := h.downloadClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		resp.Body.Close()
		return nil, fmt.Errorf("private model — set LYCHEE_HF_TOKEN env var")
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		return nil, fmt.Errorf("file not found on HuggingFace")
	}
	return resp, nil
}

func (h *hfHTTPClient) listGGUFFiles(ctx context.Context, org, repo string) ([]hfFile, error) {
	url := fmt.Sprintf("https://huggingface.co/api/models/%s/%s", org, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Lychee/1.0")
	if tok := os.Getenv("LYCHEE_HF_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := h.doAPI(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Siblings []struct {
			Rfilename string `json:"rfilename"`
			Size      int64  `json:"size"`
			LFS       *struct {
				SHA256 string `json:"sha256"`
				Size   int64  `json:"size"`
			} `json:"lfs"`
		} `json:"siblings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var files []hfFile
	for _, s := range result.Siblings {
		name := s.Rfilename
		if !strings.HasSuffix(strings.ToLower(name), ".gguf") {
			continue
		}
		f := hfFile{
			Name:     name,
			Size:     s.Size,
			quantKey: extractQuantKey(name),
		}
		if s.LFS != nil {
			f.SHA256 = s.LFS.SHA256
			if f.Size == 0 {
				f.Size = s.LFS.Size
			}
		}
		files = append(files, f)
	}
	return files, nil
}

// extractQuantKey pulls the quantization identifier from a filename
// e.g. "Llama-3.1-8B-Instruct-Q4_K_M.gguf" → "q4_k_m"
func extractQuantKey(name string) string {
	lower := strings.ToLower(strings.TrimSuffix(name, ".gguf"))
	// Handle shard patterns like "model-Q4_K_M-00001-of-00003.gguf"
	lower = strings.ReplaceAll(lower, "-of-", " ")

	quants := []string{
		"q2_k", "q3_k_s", "q3_k_m", "q3_k_l",
		"q4_0", "q4_1", "q4_k_s", "q4_k_m", "q4_k_l",
		"q5_0", "q5_1", "q5_k_s", "q5_k_m",
		"q6_k", "q8_0",
		"f16", "fp16", "f32", "bf16",
		"iq2_xs", "iq2_xxs", "iq3_xs", "iq3_xxs", "iq4_xs", "iq4_nl",
	}
	for _, q := range quants {
		if strings.Contains(lower, q) {
			return q
		}
	}
	return "other"
}

// verifySHA256 checks a file against an expected hash
func verifySHA256(path, expected string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false
	}
	return strings.EqualFold(hex.EncodeToString(h.Sum(nil)), expected)
}

// formatBytes renders a byte count as human-readable
func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// ── Search / List handlers ────────────────────────────────────────────────────

func fetchDynamicHFModels(query string) [][4]string {
	url := fmt.Sprintf("https://huggingface.co/api/models?search=%s&filter=gguf&sort=downloads&direction=-1&limit=30", query)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Lychee/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var hfResp []struct {
		ID        string   `json:"id"`
		Downloads int      `json:"downloads"`
		Likes     int      `json:"likes"`
		Tags      []string `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return nil
	}

	var entries [][4]string
	for _, item := range hfResp {
		parts := strings.Split(item.ID, "/")
		if len(parts) != 2 {
			continue
		}
		repo := parts[1]
		name := strings.ToLower(repo)
		name = strings.TrimSuffix(name, "-gguf")
		name = strings.TrimSuffix(name, ".gguf")

		tagsList := []string{"chat"}
		for _, t := range item.Tags {
			if t == "text-to-image" || t == "image-to-image" {
				tagsList = append(tagsList, "vision")
			}
		}
		tags := strings.Join(tagsList, ",")

		entries = append(entries, [4]string{
			name,
			item.ID,
			tags,
			"",
		})
	}

	return entries
}

// HFSearchHandler handles "lychee hf search <query>"
func HFSearchHandler(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(strings.Join(args, " "))
	if query == "" {
		fmt.Printf("Lychee HuggingFace Catalog — %d models\n\n", len(hfCatalogEntries()))
	} else {
		fmt.Printf("Lychee HuggingFace Catalog — searching %q\n\n", query)
	}
	fmt.Printf("  %-32s %-52s %s\n", "NAME", "HF REPO", "TAGS")
	fmt.Println("  " + strings.Repeat("-", 100))

	seen := make(map[string]bool)
	var merged [][4]string

	for _, m := range hfCatalogEntries() {
		if query == "" ||
			strings.Contains(strings.ToLower(m[0]), query) ||
			strings.Contains(strings.ToLower(m[1]), query) ||
			strings.Contains(strings.ToLower(m[2]), query) {
			merged = append(merged, m)
			seen[strings.ToLower(m[1])] = true
		}
	}

	if query != "" {
		dynamic := fetchDynamicHFModels(query)
		for _, m := range dynamic {
			repoLower := strings.ToLower(m[1])
			if !seen[repoLower] {
				merged = append(merged, m)
				seen[repoLower] = true
			}
		}
	}

	count := 0
	for _, m := range merged {
		fmt.Printf("  %-32s %-52s %s\n", m[0], m[1], m[2])
		count++
	}
	fmt.Printf("\n  %d model(s) found. Pull any with:\n    lychee hf pull <org/repo>\n    lychee hf pull <org/repo> --list   (show all quants)\n", count)
	return nil
}

// HFListHandler lists all models in the built-in catalog
func HFListHandler(cmd *cobra.Command, args []string) error {
	entries := hfCatalogEntries()
	fmt.Printf("Lychee Model Catalog — %d models, all free, no token required\n\n", len(entries))
	fmt.Printf("  %-32s %-52s %-8s %s\n", "NAME", "HF REPO", "SIZE", "TAGS")
	fmt.Println("  " + strings.Repeat("-", 108))
	for _, m := range entries {
		size := m[3]
		if size != "" {
			size += " GB"
		}
		fmt.Printf("  %-32s %-52s %-8s %s\n", m[0], m[1], size, m[2])
	}
	fmt.Printf("\n  %d models listed.\n", len(entries))
	fmt.Printf("  Pull any:  lychee hf pull <org/repo>\n")
	fmt.Printf("  List quants: lychee hf pull <org/repo> --list\n")
	fmt.Printf("  Pick quant:  lychee hf pull <org/repo> --quant q5_k_m\n")
	return nil
}
