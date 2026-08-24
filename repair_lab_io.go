package main

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type safeExtractResult struct {
	FileCount      int
	ExtractedBytes int64
	RootHint       string
}

func copyAndHashLimited(dst string, src io.Reader, limit int64) (string, int64, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", 0, err
	}
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(file, h), io.LimitReader(src, limit+1))
	if err != nil {
		return "", n, err
	}
	if n > limit {
		return "", n, fmt.Errorf("source archive exceeds the %d-byte Repair Lab intake limit", limit)
	}
	if err := file.Sync(); err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func safeExtractSourceZip(archivePath, destination string) (safeExtractResult, error) {
	var result safeExtractResult
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return result, fmt.Errorf("open source ZIP: %w", err)
	}
	defer reader.Close()
	if len(reader.File) == 0 {
		return result, errors.New("source ZIP is empty")
	}
	if len(reader.File) > repairLabMaxArchiveEntries {
		return result, fmt.Errorf("source ZIP has %d entries; the safety limit is %d", len(reader.File), repairLabMaxArchiveEntries)
	}

	tmp := destination + ".extracting-" + randomToken(6)
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return result, err
	}
	defer os.RemoveAll(tmp)

	seen := map[string]struct{}{}
	top := map[string]struct{}{}
	for _, entry := range reader.File {
		clean, err := safeArchiveEntryName(entry.Name)
		if err != nil {
			return result, err
		}
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			return result, fmt.Errorf("source ZIP contains a duplicate path: %s", clean)
		}
		seen[clean] = struct{}{}
		first := strings.Split(clean, "/")[0]
		if first != "" {
			top[first] = struct{}{}
		}
		mode := entry.Mode()
		if mode&os.ModeSymlink != 0 || mode&os.ModeType != 0 && !mode.IsDir() {
			return result, fmt.Errorf("source ZIP contains an unsupported link or special file: %s", clean)
		}
		if entry.UncompressedSize64 > uint64(repairLabMaxEntryBytes) {
			return result, fmt.Errorf("source ZIP entry exceeds the per-file safety limit: %s", clean)
		}
		if entry.CompressedSize64 > 0 && entry.UncompressedSize64 > entry.CompressedSize64*1000 && entry.UncompressedSize64 > 64<<20 {
			return result, fmt.Errorf("source ZIP entry has a suspicious compression ratio: %s", clean)
		}
		if entry.UncompressedSize64 > uint64(repairLabMaxExtractedBytes-result.ExtractedBytes) {
			return result, fmt.Errorf("source ZIP exceeds the %d-byte extracted safety limit", repairLabMaxExtractedBytes)
		}

		target := filepath.Join(tmp, filepath.FromSlash(clean))
		if !pathContainedBy(tmp, target) {
			return result, fmt.Errorf("source ZIP path escapes the staging directory: %s", clean)
		}
		if entry.FileInfo().IsDir() || strings.HasSuffix(clean, "/") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return result, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return result, err
		}
		in, err := entry.Open()
		if err != nil {
			return result, fmt.Errorf("open ZIP entry %s: %w", clean, err)
		}
		perm := mode.Perm()
		if perm == 0 {
			perm = 0o644
		}
		if isBuildWrapperName(filepath.Base(target)) {
			perm |= 0o100
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
		if err != nil {
			in.Close()
			return result, err
		}
		written, copyErr := io.Copy(out, io.LimitReader(in, int64(entry.UncompressedSize64)+1))
		closeInErr := in.Close()
		closeOutErr := out.Close()
		if copyErr != nil {
			return result, fmt.Errorf("extract %s: %w", clean, copyErr)
		}
		if closeInErr != nil {
			return result, closeInErr
		}
		if closeOutErr != nil {
			return result, closeOutErr
		}
		if written != int64(entry.UncompressedSize64) {
			return result, fmt.Errorf("source ZIP entry size mismatch for %s", clean)
		}
		result.FileCount++
		result.ExtractedBytes += written
	}
	if result.FileCount == 0 {
		return result, errors.New("source ZIP contains no regular files")
	}
	if len(top) == 1 {
		for name := range top {
			if info, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(name))); err == nil && info.IsDir() {
				result.RootHint = name
			}
		}
	}
	if err := os.RemoveAll(destination); err != nil {
		return result, err
	}
	if err := os.Rename(tmp, destination); err != nil {
		return result, err
	}
	return result, nil
}

func safeArchiveEntryName(name string) (string, error) {
	if strings.ContainsRune(name, '\x00') {
		return "", errors.New("source ZIP contains a NUL byte in a path")
	}
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimSpace(name)
	if name == "" || name == "." {
		return "", nil
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "//") || filepath.IsAbs(name) {
		return "", fmt.Errorf("source ZIP contains an absolute path: %s", name)
	}
	if len(name) >= 2 && name[1] == ':' {
		return "", fmt.Errorf("source ZIP contains a drive-qualified path: %s", name)
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	clean = strings.TrimPrefix(clean, "./")
	if clean == "." || clean == "" {
		return "", nil
	}
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("source ZIP contains a parent traversal: %s", name)
	}
	for _, component := range strings.Split(clean, "/") {
		if component == "" || component == "." || component == ".." {
			return "", fmt.Errorf("source ZIP contains an unsafe path component: %s", name)
		}
		if isWindowsDeviceName(component) {
			return "", fmt.Errorf("source ZIP contains a Windows device path: %s", name)
		}
	}
	return clean, nil
}

func isWindowsDeviceName(component string) bool {
	base := strings.TrimSpace(strings.TrimSuffix(component, filepath.Ext(component)))
	base = strings.ToUpper(strings.TrimRight(base, ". "))
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	if len(base) == 4 {
		prefix, digit := base[:3], base[3]
		return (prefix == "COM" || prefix == "LPT") && digit >= '1' && digit <= '9'
	}
	return false
}

func pathContainedBy(root, candidate string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, candidateAbs)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func isBuildWrapperName(name string) bool {
	switch strings.ToLower(name) {
	case "gradlew", "mvnw":
		return true
	default:
		return false
	}
}

func hashFileSHA256(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	h := sha256.New()
	n, err := io.Copy(h, file)
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func hashDirectoryTree(root string) (string, int, int64, error) {
	type item struct {
		rel  string
		path string
		mode fs.FileMode
		size int64
	}
	items := make([]item, 0, 512)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("immutable source contains a symlink: %s", rel)
		}
		items = append(items, item{rel: rel, path: path, mode: info.Mode(), size: info.Size()})
		return nil
	})
	if err != nil {
		return "", 0, 0, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].rel < items[j].rel })
	h := sha256.New()
	writer := bufio.NewWriterSize(hashWriter{h}, 64<<10)
	files := 0
	var bytes int64
	for _, current := range items {
		kind := "file"
		if current.mode.IsDir() {
			kind = "dir"
		}
		_, _ = fmt.Fprintf(writer, "%s\x00%s\x00%o\x00%d\n", kind, current.rel, current.mode.Perm(), current.size)
		if !current.mode.IsRegular() {
			continue
		}
		file, err := os.Open(current.path)
		if err != nil {
			return "", files, bytes, err
		}
		n, err := io.Copy(writer, file)
		file.Close()
		if err != nil {
			return "", files, bytes, err
		}
		_, _ = writer.WriteString("\n\x00\n")
		files++
		bytes += n
	}
	if err := writer.Flush(); err != nil {
		return "", files, bytes, err
	}
	return hex.EncodeToString(h.Sum(nil)), files, bytes, nil
}

type hashWriter struct{ io.Writer }

func resetWorkingCopy(run *PortingBuildRun) error {
	if run == nil {
		return errors.New("nil repair session")
	}
	if err := os.RemoveAll(run.Paths.WorkingCopy); err != nil {
		return err
	}
	if err := copyDir(run.Paths.ImmutableSource, run.Paths.WorkingCopy); err != nil {
		return err
	}
	return makeBuildWrappersExecutable(run.Paths.WorkingCopy)
}

func makeBuildWrappersExecutable(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isBuildWrapperName(entry.Name()) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return os.Chmod(path, info.Mode().Perm()|0o100)
	})
}

func verifyImmutableSource(run *PortingBuildRun) error {
	if run == nil || run.Source.TreeSHA256 == "" {
		return errors.New("immutable source identity is unavailable")
	}
	digest, files, bytes, err := hashDirectoryTree(run.Paths.ImmutableSource)
	if err != nil {
		return err
	}
	if digest != run.Source.TreeSHA256 || files != run.Source.FileCount || bytes != run.Source.ExtractedBytes {
		return fmt.Errorf("immutable source verification failed: expected %s/%d/%d, got %s/%d/%d", run.Source.TreeSHA256, run.Source.FileCount, run.Source.ExtractedBytes, digest, files, bytes)
	}
	return nil
}

func tailTextFile(path string, maxBytes int64) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return ""
	}
	start := info.Size() - maxBytes
	if start < 0 {
		start = 0
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return ""
	}
	data, _ := io.ReadAll(io.LimitReader(file, maxBytes))
	if start > 0 {
		if idx := strings.IndexByte(string(data), '\n'); idx >= 0 {
			data = data[idx+1:]
		}
	}
	return string(data)
}

func zipDirectoryDeterministic(root, output string, include func(rel string, entry fs.DirEntry) bool) (string, int64, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", 0, err
	}
	tmp := output + ".tmp-" + randomToken(6)
	_ = os.Remove(tmp)
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", 0, err
	}
	zw := zip.NewWriter(file)
	type zipItem struct {
		path string
		rel  string
		info fs.FileInfo
	}
	items := []zipItem{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if include != nil && !include(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("export refuses symlink: %s", rel)
		}
		items = append(items, zipItem{path: path, rel: rel, info: info})
		return nil
	})
	if err != nil {
		zw.Close()
		file.Close()
		os.Remove(tmp)
		return "", 0, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].rel < items[j].rel })
	fixed := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, current := range items {
		name := current.rel
		if current.info.IsDir() {
			name += "/"
		}
		header, err := zip.FileInfoHeader(current.info)
		if err != nil {
			zw.Close()
			file.Close()
			os.Remove(tmp)
			return "", 0, err
		}
		header.Name = name
		header.Modified = fixed
		header.Method = zip.Deflate
		if current.info.IsDir() {
			header.Method = zip.Store
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			zw.Close()
			file.Close()
			os.Remove(tmp)
			return "", 0, err
		}
		if current.info.IsDir() {
			continue
		}
		in, err := os.Open(current.path)
		if err != nil {
			zw.Close()
			file.Close()
			os.Remove(tmp)
			return "", 0, err
		}
		_, copyErr := io.Copy(writer, in)
		in.Close()
		if copyErr != nil {
			zw.Close()
			file.Close()
			os.Remove(tmp)
			return "", 0, copyErr
		}
	}
	if err := zw.Close(); err != nil {
		file.Close()
		os.Remove(tmp)
		return "", 0, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tmp)
		return "", 0, err
	}
	if err := file.Close(); err != nil {
		os.Remove(tmp)
		return "", 0, err
	}
	if err := os.Rename(tmp, output); err != nil {
		os.Remove(tmp)
		return "", 0, err
	}
	return hashFileSHA256(output)
}
