package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultWhisperModelName = "large-v3-turbo-q5_0"
const defaultWhisperModelURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin?download=true"
const defaultWhisperModelSHA1 = "e050f7970618a659205450ad97eb95a18d69c9ee"

// Package variables keep the production defaults immutable from the UI while allowing
// deterministic local integration tests without depending on external DNS.
var githubReleaseAPIBase = "https://api.github.com"
var whisperModelDownloadURL = defaultWhisperModelURL
var whisperModelExpectedSHA1 = defaultWhisperModelSHA1

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	Digest             string `json:"digest"`
}

type creatorToolchain struct {
	YTDLP      string
	FFmpeg     string
	WhisperCLI string
	Model      string
}

type whisperModelSpec struct {
	Name     string
	FileName string
	URL      string
	SHA1     string
	Label    string
}

func creatorWhisperModel(name string) (whisperModelSpec, bool) {
	name = strings.TrimSpace(strings.ToLower(name))
	models := map[string]whisperModelSpec{
		"large-v3-turbo-q5_0": {Name: "large-v3-turbo-q5_0", FileName: "ggml-large-v3-turbo-q5_0.bin", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin?download=true", SHA1: "e050f7970618a659205450ad97eb95a18d69c9ee", Label: "Whisper large-v3-turbo Q5 archival"},
		"large-v3-turbo":      {Name: "large-v3-turbo", FileName: "ggml-large-v3-turbo.bin", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo.bin?download=true", SHA1: "4af2b29d7ec73d781377bfd1758ca957a807e941", Label: "Whisper large-v3-turbo archival"},
		"base":                {Name: "base", FileName: "ggml-base.bin", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin?download=true", SHA1: "465707469ff3a37a2b9b8d8f89f2f99de7299dac", Label: "Whisper base multilingual"},
	}
	m, ok := models[name]
	if ok && name == defaultWhisperModelName {
		m.URL = whisperModelDownloadURL
		m.SHA1 = whisperModelExpectedSHA1
	}
	return m, ok
}

func (a *App) selectedCreatorWhisperModel() whisperModelSpec {
	a.mu.RLock()
	name := strings.TrimSpace(a.settings.CreatorTranscriptModel)
	a.mu.RUnlock()
	if name == "" {
		name = defaultWhisperModelName
	}
	if m, ok := creatorWhisperModel(name); ok {
		return m
	}
	m, _ := creatorWhisperModel(defaultWhisperModelName)
	return m
}

func (a *App) localCreatorTranscript(ctx context.Context, rawURL string) ([]TranscriptSegment, string, error) {
	tools, err := a.ensureCreatorTranscriptionToolchain(ctx)
	if err != nil {
		return nil, "", err
	}
	workRoot := filepath.Join(a.cfgDir, "transcription-work")
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return nil, "", err
	}
	work, err := os.MkdirTemp(workRoot, "creator-")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(work)

	outputTemplate := filepath.Join(work, "source.%(ext)s")
	args := []string{
		"--no-playlist", "--no-warnings", "--no-progress",
		"-f", "bestaudio/best", "-x", "--audio-format", "wav", "--audio-quality", "0",
		"--ffmpeg-location", filepath.Dir(tools.FFmpeg),
		"-o", outputTemplate,
		rawURL,
	}
	cmd := exec.CommandContext(ctx, tools.YTDLP, args...)
	cmd.Env = append(os.Environ(), "PYTHONUTF8=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("video audio extraction failed: %w (%s)", err, truncate(strings.TrimSpace(string(out)), 600))
	}
	wav, err := findFileByExt(work, ".wav")
	if err != nil {
		return nil, "", err
	}

	outBase := filepath.Join(work, "transcript")
	whisperArgs := []string{
		"-m", tools.Model,
		"-f", wav,
		"-osrt", "-of", outBase,
		"-l", "auto", "-np",
		"--prompt", "Minecraft mods, Modrinth, CurseForge, Fabric, Forge, NeoForge, Quilt, mod loader, resource pack, shader, addon",
	}
	whisperCmd := exec.CommandContext(ctx, tools.WhisperCLI, whisperArgs...)
	whisperCmd.Dir = filepath.Dir(tools.WhisperCLI)
	if out, err := whisperCmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("local Whisper transcription failed: %w (%s)", err, truncate(strings.TrimSpace(string(out)), 600))
	}
	b, err := os.ReadFile(outBase + ".srt")
	if err != nil {
		return nil, "", fmt.Errorf("Whisper produced no SRT transcript: %w", err)
	}
	segments := parseTimedText(string(b))
	if len(segments) == 0 {
		return nil, "", errors.New("local Whisper transcript contained no timed speech segments")
	}
	return segments, "local Whisper.cpp ASR", nil
}

func (a *App) ensureCreatorYTDLP(ctx context.Context) (string, error) {
	root := filepath.Join(a.cfgDir, "creator-transcription")
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return "", err
	}
	ytdlpName := "yt-dlp_linux"
	if runtime.GOOS == "windows" {
		ytdlpName = "yt-dlp.exe"
	}
	ytdlpPath := filepath.Join(toolsDir, ytdlpName)
	if pathExists(ytdlpPath) {
		return ytdlpPath, nil
	}
	asset, err := a.latestGitHubReleaseAsset(ctx, "yt-dlp/yt-dlp", func(name string) bool { return name == ytdlpName })
	if err != nil {
		return "", fmt.Errorf("yt-dlp bootstrap: %w", err)
	}
	if err := a.downloadReleaseAsset(ctx, asset, ytdlpPath); err != nil {
		return "", fmt.Errorf("yt-dlp bootstrap: %w", err)
	}
	_ = os.Chmod(ytdlpPath, 0o755)
	return ytdlpPath, nil
}

func (a *App) ytDLPTranscript(ctx context.Context, rawURL string) ([]TranscriptSegment, string, error) {
	ytdlp, err := a.ensureCreatorYTDLP(ctx)
	if err != nil {
		return nil, "", err
	}
	workRoot := filepath.Join(a.cfgDir, "transcription-work")
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return nil, "", err
	}
	work, err := os.MkdirTemp(workRoot, "captions-")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(work)
	outTemplate := filepath.Join(work, "captions.%(ext)s")
	args := []string{"--no-playlist", "--no-warnings", "--no-progress", "--skip-download", "--write-subs", "--write-auto-subs", "--sub-langs", "en.*,en", "--sub-format", "vtt", "-o", outTemplate, rawURL}
	cmd := exec.CommandContext(ctx, ytdlp, args...)
	cmd.Env = append(os.Environ(), "PYTHONUTF8=1")
	if out, e := cmd.CombinedOutput(); e != nil {
		return nil, "", fmt.Errorf("yt-dlp caption extraction failed: %w (%s)", e, truncate(strings.TrimSpace(string(out)), 500))
	}
	vtt, e := findFileByExt(work, ".vtt")
	if e != nil {
		return nil, "", errors.New("yt-dlp found no public or automatic subtitle track")
	}
	b, e := os.ReadFile(vtt)
	if e != nil {
		return nil, "", e
	}
	segs := parseTimedText(string(b))
	if len(segs) == 0 {
		return nil, "", errors.New("yt-dlp subtitle track contained no timed speech segments")
	}
	return segs, "yt-dlp public/automatic captions", nil
}

type CreatorTranscriptRecord struct {
	VideoID      string                  `json:"videoId"`
	Source       string                  `json:"source"`
	SavedAt      string                  `json:"savedAt"`
	SegmentCount int                     `json:"segmentCount"`
	WordCount    int                     `json:"wordCount"`
	Segments     []TranscriptSegmentJSON `json:"segments"`
}
type TranscriptSegmentJSON struct {
	StartMS   int64  `json:"startMs"`
	EndMS     int64  `json:"endMs"`
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
	Source    string `json:"source,omitempty"`
}

func (a *App) saveCreatorTranscript(videoID, source string, segments []TranscriptSegment) error {
	if videoID == "" || len(segments) == 0 {
		return nil
	}
	rec := CreatorTranscriptRecord{VideoID: videoID, Source: source, SavedAt: time.Now().UTC().Format(time.RFC3339), SegmentCount: len(segments)}
	for _, seg := range segments {
		rec.WordCount += len(strings.Fields(seg.Text))
		rec.Segments = append(rec.Segments, TranscriptSegmentJSON{StartMS: seg.StartMS, EndMS: seg.EndMS, Timestamp: formatTimestamp(seg.StartMS / 1000), Text: seg.Text, Source: seg.Source})
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(a.cfgDir, "creator-transcripts", videoID+".json"), b, 0o644)
}
func (a *App) loadCreatorTranscript(videoID string) (CreatorTranscriptRecord, error) {
	var rec CreatorTranscriptRecord
	b, err := os.ReadFile(filepath.Join(a.cfgDir, "creator-transcripts", videoID+".json"))
	if err != nil {
		return rec, err
	}
	err = json.Unmarshal(b, &rec)
	return rec, err
}

func (a *App) ensureCreatorTranscriptionToolchain(ctx context.Context) (creatorToolchain, error) {
	root := filepath.Join(a.cfgDir, "creator-transcription")
	toolsDir := filepath.Join(root, "tools")
	modelDir := filepath.Join(root, "models")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return creatorToolchain{}, err
	}
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		return creatorToolchain{}, err
	}

	ytdlpPath, err := a.ensureCreatorYTDLP(ctx)
	if err != nil {
		return creatorToolchain{}, err
	}

	whisperRoot := filepath.Join(toolsDir, "whisper.cpp")
	whisperName := "whisper-cli"
	assetName := "whisper-bin-ubuntu-x64.tar.gz"
	if runtime.GOOS == "windows" {
		whisperName = "whisper-cli.exe"
		assetName = "whisper-bin-x64.zip"
	}
	whisperPath, _ := findFileByBase(whisperRoot, whisperName)
	if whisperPath == "" {
		asset, err := a.latestGitHubReleaseAsset(ctx, "ggml-org/whisper.cpp", func(name string) bool { return name == assetName })
		if err != nil {
			return creatorToolchain{}, fmt.Errorf("whisper.cpp bootstrap: %w", err)
		}
		archive := filepath.Join(toolsDir, asset.Name)
		if err := a.downloadReleaseAsset(ctx, asset, archive); err != nil {
			return creatorToolchain{}, fmt.Errorf("whisper.cpp bootstrap: %w", err)
		}
		_ = os.RemoveAll(whisperRoot)
		if err := os.MkdirAll(whisperRoot, 0o755); err != nil {
			return creatorToolchain{}, err
		}
		if strings.HasSuffix(strings.ToLower(asset.Name), ".zip") {
			err = extractZipSafe(archive, whisperRoot)
		} else {
			err = extractTarGzSafe(archive, whisperRoot)
		}
		_ = os.Remove(archive)
		if err != nil {
			return creatorToolchain{}, fmt.Errorf("whisper.cpp extraction: %w", err)
		}
		whisperPath, err = findFileByBase(whisperRoot, whisperName)
		if err != nil {
			return creatorToolchain{}, fmt.Errorf("whisper.cpp executable: %w", err)
		}
		_ = os.Chmod(whisperPath, 0o755)
	}

	ffmpegPath, err := a.ensureCreatorFFmpeg(ctx)
	if err != nil {
		return creatorToolchain{}, err
	}

	modelName := defaultWhisperModelName
	a.mu.RLock()
	if strings.TrimSpace(a.settings.CreatorTranscriptModel) != "" {
		modelName = strings.TrimSpace(a.settings.CreatorTranscriptModel)
	}
	a.mu.RUnlock()
	modelSpec, ok := creatorWhisperModel(modelName)
	if !ok {
		return creatorToolchain{}, fmt.Errorf("unsupported creator transcript model %q", modelName)
	}
	modelPath := filepath.Join(modelDir, modelSpec.FileName)
	if !pathExists(modelPath) {
		part := modelPath + ".part"
		if err := a.downloadURLVerified(ctx, modelSpec.URL, part, 0, map[string]string{"sha1": modelSpec.SHA1}); err != nil {
			_ = os.Remove(part)
			return creatorToolchain{}, fmt.Errorf("Whisper model bootstrap: %w", err)
		}
		if err := os.Rename(part, modelPath); err != nil {
			return creatorToolchain{}, err
		}
	}
	return creatorToolchain{YTDLP: ytdlpPath, FFmpeg: ffmpegPath, WhisperCLI: whisperPath, Model: modelPath}, nil
}

func (a *App) ensureCreatorFFmpeg(ctx context.Context) (string, error) {
	if ffmpegPath, _ := exec.LookPath("ffmpeg"); ffmpegPath != "" {
		return ffmpegPath, nil
	}
	if runtime.GOOS != "windows" {
		return "", errors.New("FFmpeg is required for creator video analysis and was not found on this platform")
	}
	toolsDir := filepath.Join(a.cfgDir, "creator-transcription", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return "", err
	}
	ffmpegRoot := filepath.Join(toolsDir, "ffmpeg")
	if ffmpegPath, _ := findFileByBase(ffmpegRoot, "ffmpeg.exe"); ffmpegPath != "" {
		return ffmpegPath, nil
	}
	asset, err := a.latestGitHubReleaseAsset(ctx, "BtbN/FFmpeg-Builds", func(name string) bool {
		low := strings.ToLower(name)
		return strings.HasSuffix(low, "win64-gpl.zip") && !strings.Contains(low, "shared")
	})
	if err != nil {
		return "", fmt.Errorf("FFmpeg bootstrap: %w", err)
	}
	archive := filepath.Join(toolsDir, asset.Name)
	if err := a.downloadReleaseAsset(ctx, asset, archive); err != nil {
		return "", fmt.Errorf("FFmpeg bootstrap: %w", err)
	}
	_ = os.RemoveAll(ffmpegRoot)
	if err := os.MkdirAll(ffmpegRoot, 0o755); err != nil {
		return "", err
	}
	if err := extractZipSafe(archive, ffmpegRoot); err != nil {
		return "", fmt.Errorf("FFmpeg extraction: %w", err)
	}
	_ = os.Remove(archive)
	ffmpegPath, err := findFileByBase(ffmpegRoot, "ffmpeg.exe")
	if err != nil {
		return "", fmt.Errorf("FFmpeg executable: %w", err)
	}
	return ffmpegPath, nil
}

func (a *App) latestGitHubReleaseAsset(ctx context.Context, repo string, match func(string) bool) (githubReleaseAsset, error) {
	a.mu.RLock()
	token := strings.TrimSpace(a.settings.GitHubToken)
	a.mu.RUnlock()
	headers := map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	var rel struct {
		TagName string               `json:"tag_name"`
		Assets  []githubReleaseAsset `json:"assets"`
	}
	if err := a.getJSON(ctx, strings.TrimRight(githubReleaseAPIBase, "/")+"/repos/"+repo+"/releases/latest", headers, &rel); err != nil {
		return githubReleaseAsset{}, err
	}
	for _, asset := range rel.Assets {
		if match(asset.Name) && asset.BrowserDownloadURL != "" {
			return asset, nil
		}
	}
	return githubReleaseAsset{}, fmt.Errorf("release %s has no matching asset", rel.TagName)
}

func (a *App) downloadReleaseAsset(ctx context.Context, asset githubReleaseAsset, dst string) error {
	if asset.BrowserDownloadURL == "" {
		return errors.New("release asset has no download URL")
	}
	hashes := map[string]string{}
	if strings.HasPrefix(strings.ToLower(asset.Digest), "sha256:") {
		hashes["sha256"] = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(asset.Digest), "sha256:"))
	}
	part := dst + ".part"
	_ = os.Remove(part)
	if err := a.downloadURLVerified(ctx, asset.BrowserDownloadURL, part, asset.Size, hashes); err != nil {
		_ = os.Remove(part)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		_ = os.Remove(part)
		return err
	}
	_ = os.Remove(dst)
	return os.Rename(part, dst)
}

func extractZipSafe(src, dst string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		target, err := archiveTarget(dst, f.Name)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			r.Close()
			return err
		}
		_, copyErr := io.Copy(out, io.LimitReader(r, 2<<30))
		closeErr := out.Close()
		r.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func extractTarGzSafe(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		target, err := archiveTarget(dst, h.Name)
		if err != nil {
			return err
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, io.LimitReader(tr, 2<<30))
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
	}
	return nil
}

func archiveTarget(root, name string) (string, error) {
	name = filepath.FromSlash(name)
	clean := filepath.Clean(name)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", errors.New("archive contains an unsafe path")
	}
	target := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("archive path escaped destination")
	}
	return target, nil
}

func findFileByBase(root, base string) (string, error) {
	if root == "" || !pathExists(root) {
		return "", os.ErrNotExist
	}
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.EqualFold(d.Name(), base) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if found == "" {
		if err != nil {
			return "", err
		}
		return "", os.ErrNotExist
	}
	return found, nil
}

func findFileByExt(root, ext string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), ext) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if found == "" {
		if err != nil {
			return "", err
		}
		return "", os.ErrNotExist
	}
	return found, nil
}

func (a *App) handleTranscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	root := filepath.Join(a.cfgDir, "creator-transcription")
	spec := a.selectedCreatorWhisperModel()
	model := filepath.Join(root, "models", spec.FileName)
	ytdlp := filepath.Join(root, "tools", "yt-dlp_linux")
	if runtime.GOOS == "windows" {
		ytdlp = filepath.Join(root, "tools", "yt-dlp.exe")
	}
	whisperName := "whisper-cli"
	if runtime.GOOS == "windows" {
		whisperName += ".exe"
	}
	whisper, _ := findFileByBase(filepath.Join(root, "tools", "whisper.cpp"), whisperName)
	ffmpeg, _ := exec.LookPath("ffmpeg")
	if runtime.GOOS == "windows" && ffmpeg == "" {
		ffmpeg, _ = findFileByBase(filepath.Join(root, "tools", "ffmpeg"), "ffmpeg.exe")
	}
	ocr, ocrErr := detectCreatorOCRBackend()
	writeJSON(w, http.StatusOK, map[string]any{
		"ready": pathExists(model) && pathExists(ytdlp) && whisper != "" && ffmpeg != "",
		"ytDlp": pathExists(ytdlp), "whisper": whisper != "", "model": pathExists(model), "ffmpeg": ffmpeg != "",
		"visualOcr": ocrErr == nil, "visualOcrBackend": ocr.Name,
		"modelName": spec.Label, "modelKey": spec.Name, "autoBootstrap": true,
	})
}

func (a *App) handleTranscriptionPrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Minute)
	defer cancel()
	tools, err := a.ensureCreatorTranscriptionToolchain(ctx)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	spec := a.selectedCreatorWhisperModel()
	ocr, ocrErr := detectCreatorOCRBackend()
	writeJSON(w, http.StatusOK, map[string]any{
		"ready": true, "ytDlp": pathExists(tools.YTDLP), "whisper": pathExists(tools.WhisperCLI),
		"model": pathExists(tools.Model), "ffmpeg": pathExists(tools.FFmpeg),
		"visualOcr": ocrErr == nil, "visualOcrBackend": ocr.Name,
		"modelName": spec.Label, "modelKey": spec.Name, "autoBootstrap": true,
	})
}

// Keep the compiler honest about the release digest parser if GitHub changes shape.
func releaseAssetDigestSHA256(asset githubReleaseAsset) string {
	if !strings.HasPrefix(strings.ToLower(asset.Digest), "sha256:") {
		return ""
	}
	h := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(asset.Digest), "sha256:"))
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != sha256.Size {
		return ""
	}
	return hex.EncodeToString(b)
}
