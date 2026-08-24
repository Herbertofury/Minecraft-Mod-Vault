package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func providerSupportsVerifiedDetectedInstall(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "curseforge", "planetminecraft", "mcpedl", "bukkitdev", "moddb", "minecraftmaps", "resourcepacknet", "texturepacks", "mcreator", "shaderpackscom", "shaderpacksnet", "minecraftshader":
		return true
	default:
		return false
	}
}

func (a *App) installDetectedWebPackage(ctx context.Context, provider, pageURL, id, target string) (map[string]any, error) {
	if !providerSupportsVerifiedDetectedInstall(provider) {
		return nil, errors.New("provider does not use verified detected-package installation")
	}
	if strings.TrimSpace(pageURL) == "" {
		return nil, errors.New("project page is required for detected-package installation")
	}
	details, err := a.genericWebDetails(ctx, provider, pageURL, id)
	if err != nil {
		return nil, err
	}
	if len(details.Links) == 0 {
		return nil, errors.New("the integrated project page did not expose a downloadable package candidate")
	}
	if target == "" || target == "auto" {
		switch details.ProjectType {
		case "resourcepack":
			target = "resourcepacks"
		case "shader":
			target = "shaderpacks"
		case "datapack":
			target = "datapacks"
		case "plugin":
			target = "plugins"
		case "mod":
			target = "mods"
		case "world":
			target = "worlds"
		default:
			target = "downloads"
		}
	}
	if !validImportTarget(target) {
		return nil, fmt.Errorf("unsupported install target %q", target)
	}

	type detectedCandidate struct {
		Label string
		URL   string
		Depth int
	}
	labels := make([]string, 0, len(details.Links))
	for label := range details.Links {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	queue := make([]detectedCandidate, 0, len(labels)+8)
	for _, label := range labels {
		queue = append(queue, detectedCandidate{Label: label, URL: strings.TrimSpace(details.Links[label])})
	}
	stageDir := filepath.Join(a.cfgDir, "staging", "detected", safeFilename(provider))
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return nil, err
	}
	seenCandidates := map[string]bool{}
	var failures []string
	for len(queue) > 0 && len(seenCandidates) < 24 {
		c := queue[0]
		queue = queue[1:]
		candidate := strings.TrimSpace(c.URL)
		if candidate == "" || seenCandidates[candidate] {
			continue
		}
		seenCandidates[candidate] = true
		u, parseErr := url.Parse(candidate)
		if parseErr != nil || !strings.EqualFold(u.Scheme, "https") || u.Hostname() == "" {
			failures = append(failures, c.Label+": unsafe or non-HTTPS URL")
			continue
		}
		name := detectedPackageFilename(u, details.Title)
		stage := uniquePath(filepath.Join(stageDir, name+".part"))
		if err := a.downloadURLVerified(ctx, candidate, stage, 0, nil); err != nil {
			failures = append(failures, c.Label+": "+err.Error())
			continue
		}
		if err := validateZipContainer(stage); err != nil {
			// Many major Minecraft sites use one or two first-party/intermediate
			// download pages before the actual CDN asset. Parse those pages inside
			// Vault and keep walking the download chain, but cap both depth and the
			// candidate count. Nothing is installed until a real archive validates.
			if c.Depth < 2 {
				if page, readErr := readSmallTextFile(stage, 2<<20); readErr == nil && looksLikeHTML(page) {
					for i, nested := range downloadLinks(page, candidate, 20) {
						if !seenCandidates[nested] {
							queue = append(queue, detectedCandidate{Label: fmt.Sprintf("%s > link %d", c.Label, i+1), URL: nested, Depth: c.Depth + 1})
						}
					}
				}
			}
			_ = os.Remove(stage)
			failures = append(failures, c.Label+": response was not a valid Minecraft ZIP/JAR package")
			continue
		}
		finalName := strings.TrimSuffix(filepath.Base(stage), ".part")
		ext := strings.ToLower(filepath.Ext(finalName))
		bedrockHint := ext == ".mcpack" || ext == ".mcaddon" || ext == ".mcworld" || strings.Contains(strings.ToLower(details.Title+" "+details.Summary+" "+details.Description+" "+details.PageURL), "bedrock") || strings.Contains(strings.ToLower(details.PageURL), "/minecraft-bedrock/")
		effectiveTarget := target
		if effectiveTarget == "worlds" && bedrockHint {
			effectiveTarget = "downloads"
		}
		if effectiveTarget == "worlds" {
			dst, installErr := a.installVerifiedWorldArchive(stage, details.Title)
			if installErr != nil {
				_ = os.Remove(stage)
				failures = append(failures, c.Label+": "+installErr.Error())
				continue
			}
			return map[string]any{
				"ok": true, "provider": provider, "project": details.Title, "path": dst,
				"file": filepath.Base(dst), "target": "worlds", "sourceCandidate": c.Label, "opened": false,
				"verification": "archive validated, path-safe extracted, and level.dat verified before world installation",
			}, nil
		}
		dstDir := a.javaTargetDir(effectiveTarget)
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			_ = os.Remove(stage)
			return nil, err
		}
		dst := uniquePath(filepath.Join(dstDir, finalName))
		if err := moveFilePortable(stage, dst); err != nil {
			_ = os.Remove(stage)
			return nil, err
		}
		opened := false
		if effectiveTarget == "downloads" && (ext == ".mcpack" || ext == ".mcaddon" || ext == ".mcworld") {
			if openExternal(dst) == nil {
				opened = true
			}
		}
		return map[string]any{
			"ok": true, "provider": provider, "project": details.Title, "path": dst,
			"file": filepath.Base(dst), "target": effectiveTarget, "sourceCandidate": c.Label, "opened": opened,
			"verification": "download chain resolved in-app and final archive passed ZIP container validation before installation",
		}, nil
	}
	if len(failures) > 5 {
		failures = failures[:5]
	}
	return nil, fmt.Errorf("no detected download candidate produced a valid Minecraft package: %s", strings.Join(failures, "; "))
}

func readSmallTextFile(path string, maxBytes int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(b)) > maxBytes {
		return "", errors.New("intermediate page exceeded safe parse limit")
	}
	return string(b), nil
}

func looksLikeHTML(s string) bool {
	low := strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(low, "<!doctype html") || strings.HasPrefix(low, "<html") || strings.Contains(low, "<body") || strings.Contains(low, "<a ")
}

func detectedPackageFilename(u *url.URL, title string) string {
	name := safeFilename(filepath.Base(strings.TrimRight(u.Path, "/")))
	if name == "" || name == "." || !hasMinecraftPackageExtension(name) {
		name = safeFilename(strings.TrimSpace(title))
		if name == "" || name == "package.bin" {
			name = "minecraft-package"
		}
		name += ".zip"
	}
	return name
}

func hasMinecraftPackageExtension(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jar", ".zip", ".mrpack", ".mcpack", ".mcaddon", ".mcworld":
		return true
	default:
		return false
	}
}

func moveFilePortable(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return closeErr
	}
	return os.Remove(src)
}

func (a *App) installVerifiedWorldArchive(archivePath, title string) (string, error) {
	worldsDir := a.javaTargetDir("worlds")
	if err := os.MkdirAll(worldsDir, 0o755); err != nil {
		return "", err
	}
	stageRoot, err := os.MkdirTemp(worldsDir, ".mmv-world-stage-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stageRoot)
	unpack := filepath.Join(stageRoot, "unpack")
	if err := os.MkdirAll(unpack, 0o755); err != nil {
		return "", err
	}
	if err := extractZipPathSafe(archivePath, unpack); err != nil {
		return "", err
	}
	worldRoot, err := findMinecraftWorldRoot(unpack)
	if err != nil {
		return "", err
	}
	name := safeWorldDirectoryName(title)
	dst := uniqueDirectoryPath(filepath.Join(worldsDir, name))
	if err := os.Rename(worldRoot, dst); err != nil {
		return "", fmt.Errorf("install world: %w", err)
	}
	_ = os.Remove(archivePath)
	return dst, nil
}

func extractZipPathSafe(archivePath, destination string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open world archive: %w", err)
	}
	defer zr.Close()
	if len(zr.File) == 0 || len(zr.File) > 200000 {
		return errors.New("world archive has an invalid file count")
	}
	var total uint64
	const maxWorldBytes = uint64(8 << 30)
	for _, entry := range zr.File {
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("world archive contains a symbolic link: %s", entry.Name)
		}
		clean := filepath.Clean(filepath.FromSlash(entry.Name))
		if clean == "." {
			continue
		}
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("world archive contains an unsafe path: %s", entry.Name)
		}
		total += entry.UncompressedSize64
		if total > maxWorldBytes {
			return errors.New("world archive expands beyond the 8 GiB safety limit")
		}
		dst := filepath.Join(destination, clean)
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		rc, err := entry.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, io.LimitReader(rc, int64(entry.UncompressedSize64)+1))
		closeErr := out.Close()
		rcErr := rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if rcErr != nil {
			return rcErr
		}
	}
	return nil
}

func findMinecraftWorldRoot(root string) (string, error) {
	best := ""
	bestDepth := int(^uint(0) >> 1)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.EqualFold(entry.Name(), "level.dat") {
			return nil
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		depth := 0
		if rel != "." {
			depth = len(strings.Split(filepath.ToSlash(rel), "/"))
		}
		if depth < bestDepth {
			best, bestDepth = filepath.Dir(path), depth
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if best == "" {
		return "", errors.New("map archive did not contain a Minecraft level.dat")
	}
	return best, nil
}

func safeWorldDirectoryName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\\|?*`, r) {
			return '_'
		}
		return r
	}, name)
	name = strings.Trim(name, " .")
	if name == "" {
		name = "Minecraft Mod Vault World"
	}
	upper := strings.ToUpper(name)
	for _, reserved := range []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"} {
		if upper == reserved {
			name = "_" + name
			break
		}
	}
	return name
}

func uniqueDirectoryPath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	for i := 2; i < 10000; i++ {
		candidate := fmt.Sprintf("%s (%d)", path, i)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
	return path + "-" + randomToken(4)
}
