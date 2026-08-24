package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type creatorOCRBackend struct {
	Name    string
	Command string
	Kind    string
}

func detectCreatorOCRBackend() (creatorOCRBackend, error) {
	if runtime.GOOS == "windows" {
		for _, name := range []string{"powershell.exe", "powershell"} {
			if p, err := exec.LookPath(name); err == nil {
				return creatorOCRBackend{Name: "Windows.Media.Ocr", Command: p, Kind: "windows"}, nil
			}
		}
	}
	if p, err := exec.LookPath("tesseract"); err == nil {
		return creatorOCRBackend{Name: "Tesseract OCR", Command: p, Kind: "tesseract"}, nil
	}
	return creatorOCRBackend{}, errors.New("no local visual OCR backend is available")
}

func (a *App) creatorVisualText(ctx context.Context, rawURL string) ([]TranscriptSegment, string, error) {
	backend, err := detectCreatorOCRBackend()
	if err != nil {
		return nil, "", err
	}
	ytdlp, err := a.ensureCreatorYTDLP(ctx)
	if err != nil {
		return nil, "", err
	}
	ffmpeg, err := a.ensureCreatorFFmpeg(ctx)
	if err != nil {
		return nil, "", err
	}
	workRoot := filepath.Join(a.cfgDir, "transcription-work")
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return nil, "", err
	}
	work, err := os.MkdirTemp(workRoot, "visual-")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(work)

	outputTemplate := filepath.Join(work, "source.%(ext)s")
	args := []string{
		"--no-playlist", "--no-warnings", "--no-progress",
		"-f", "bestvideo*+bestaudio/best",
		"--merge-output-format", "mp4",
		"--ffmpeg-location", filepath.Dir(ffmpeg),
		"-o", outputTemplate,
		rawURL,
	}
	cmd := exec.CommandContext(ctx, ytdlp, args...)
	cmd.Env = append(os.Environ(), "PYTHONUTF8=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("visual-text video acquisition failed: %w (%s)", err, truncate(strings.TrimSpace(string(out)), 600))
	}
	videoPath, err := findCreatorVideoFile(work)
	if err != nil {
		return nil, "", err
	}

	framePattern := filepath.Join(work, "frame-%04d.png")
	frameArgs := []string{
		"-hide_banner", "-loglevel", "error", "-i", videoPath,
		"-vf", "fps=1,scale='min(1600,iw)':-2",
		"-frames:v", "180",
		framePattern,
	}
	if out, err := exec.CommandContext(ctx, ffmpeg, frameArgs...).CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("visual-text frame extraction failed: %w (%s)", err, truncate(strings.TrimSpace(string(out)), 600))
	}
	frames, err := filepath.Glob(filepath.Join(work, "frame-*.png"))
	if err != nil || len(frames) == 0 {
		return nil, "", errors.New("visual-text frame extraction produced no frames")
	}
	sort.Strings(frames)

	segments := make([]TranscriptSegment, 0, len(frames))
	seen := map[string]bool{}
	for i, frame := range frames {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
		}
		text, err := a.ocrCreatorFrame(ctx, backend, frame)
		if err != nil {
			continue
		}
		text = cleanCreatorOCRText(text)
		sig := creatorTextSignature(text)
		if sig == "" || seen[sig] {
			continue
		}
		seen[sig] = true
		start := int64(i * 1000)
		segments = append(segments, TranscriptSegment{StartMS: start, EndMS: start + 1000, Text: text, Source: "visual-ocr"})
	}
	if len(segments) == 0 {
		return nil, "", errors.New("visual OCR found no readable on-screen text")
	}
	return segments, "visual OCR (" + backend.Name + ")", nil
}

func findCreatorVideoFile(root string) (string, error) {
	allowed := map[string]bool{".mp4": true, ".webm": true, ".mkv": true, ".mov": true, ".m4v": true}
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !allowed[strings.ToLower(filepath.Ext(entry.Name()))] {
			continue
		}
		return filepath.Join(root, entry.Name()), nil
	}
	return "", errors.New("visual-text analysis could not locate the downloaded video")
}

func (a *App) ocrCreatorFrame(ctx context.Context, backend creatorOCRBackend, frame string) (string, error) {
	switch backend.Kind {
	case "tesseract":
		cmd := exec.CommandContext(ctx, backend.Command, frame, "stdout", "-l", "eng", "--psm", "6")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	case "windows":
		script, err := a.ensureWindowsCreatorOCRScript()
		if err != nil {
			return "", err
		}
		cmd := exec.CommandContext(ctx, backend.Command, "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", script, frame)
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	default:
		return "", errors.New("unsupported visual OCR backend")
	}
}

const windowsCreatorOCRScript = `param([Parameter(Mandatory=$true)][string]$Path)
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$asTaskGeneric = ([System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Count -eq 1 })[0]
function Await($WinRtTask, $ResultType) {
    $asTask = $asTaskGeneric.MakeGenericMethod($ResultType)
    $netTask = $asTask.Invoke($null, @($WinRtTask))
    $netTask.Wait(-1) | Out-Null
    return $netTask.Result
}
[Windows.Storage.StorageFile, Windows.Storage, ContentType=WindowsRuntime] | Out-Null
[Windows.Storage.Streams.IRandomAccessStream, Windows.Storage.Streams, ContentType=WindowsRuntime] | Out-Null
[Windows.Graphics.Imaging.BitmapDecoder, Windows.Graphics.Imaging, ContentType=WindowsRuntime] | Out-Null
[Windows.Graphics.Imaging.SoftwareBitmap, Windows.Graphics.Imaging, ContentType=WindowsRuntime] | Out-Null
[Windows.Media.Ocr.OcrEngine, Windows.Media.Ocr, ContentType=WindowsRuntime] | Out-Null
[Windows.Media.Ocr.OcrResult, Windows.Media.Ocr, ContentType=WindowsRuntime] | Out-Null
$file = Await ([Windows.Storage.StorageFile]::GetFileFromPathAsync($Path)) ([Windows.Storage.StorageFile])
$stream = Await ($file.OpenAsync([Windows.Storage.FileAccessMode]::Read)) ([Windows.Storage.Streams.IRandomAccessStream])
$decoder = Await ([Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($stream)) ([Windows.Graphics.Imaging.BitmapDecoder])
$bitmap = Await ($decoder.GetSoftwareBitmapAsync()) ([Windows.Graphics.Imaging.SoftwareBitmap])
$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
if ($null -eq $engine) { throw 'Windows OCR engine is unavailable for the current language profile.' }
$result = Await ($engine.RecognizeAsync($bitmap)) ([Windows.Media.Ocr.OcrResult])
$result.Text
`

func (a *App) ensureWindowsCreatorOCRScript() (string, error) {
	path := filepath.Join(a.cfgDir, "creator-transcription", "tools", "windows-media-ocr.ps1")
	if pathExists(path) {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := writeFileAtomic(path, []byte(windowsCreatorOCRScript), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

var creatorOCRWhitespace = regexp.MustCompile(`[\t\r ]+`)
var creatorOCRBlankLines = regexp.MustCompile(`\n{3,}`)
var creatorOCRSignatureNoise = regexp.MustCompile(`[^a-z0-9+#._-]+`)

func cleanCreatorOCRText(raw string) string {
	raw = strings.ReplaceAll(raw, "\ufeff", "")
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(creatorOCRWhitespace.ReplaceAllString(line, " "))
		if line == "" {
			continue
		}
		clean = append(clean, line)
	}
	return strings.TrimSpace(creatorOCRBlankLines.ReplaceAllString(strings.Join(clean, "\n"), "\n\n"))
}

func creatorTextSignature(raw string) string {
	v := strings.ToLower(cleanCreatorOCRText(raw))
	v = creatorOCRSignatureNoise.ReplaceAllString(v, " ")
	return strings.Join(strings.Fields(v), " ")
}

func mergeTranscriptSegments(groups ...[]TranscriptSegment) []TranscriptSegment {
	all := []TranscriptSegment{}
	for _, group := range groups {
		all = append(all, group...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].StartMS == all[j].StartMS {
			return all[i].EndMS < all[j].EndMS
		}
		return all[i].StartMS < all[j].StartMS
	})
	out := make([]TranscriptSegment, 0, len(all))
	seen := map[string]bool{}
	for _, seg := range all {
		seg.Text = cleanCreatorOCRText(seg.Text)
		sig := creatorTextSignature(seg.Text)
		if sig == "" || seen[sig] {
			continue
		}
		seen[sig] = true
		out = append(out, seg)
	}
	return out
}

func combineCreatorTranscriptSource(existing, added string) string {
	parts := []string{}
	seen := map[string]bool{}
	for _, raw := range []string{existing, added} {
		for _, part := range strings.Split(raw, " + ") {
			part = strings.TrimSpace(part)
			key := strings.ToLower(part)
			if part == "" || seen[key] {
				continue
			}
			seen[key] = true
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " + ")
}
