package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func snapshotPath(label, path string) (ContentSnapshot, error) {
	st, err := os.Stat(path)
	if err != nil {
		return ContentSnapshot{}, err
	}
	if st.IsDir() {
		return snapshotDirectory(label, path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ContentSnapshot{}, err
	}
	return snapshotBytes(label, path, filepath.Base(path), b)
}

func snapshotDirectory(label, root string) (ContentSnapshot, error) {
	out := ContentSnapshot{Label: label, Source: root, Categories: map[string]int{}}
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		ent := makeEntry(rel, b)
		out.Entries = append(out.Entries, ent)
		out.Categories[ent.Category]++
		_, _ = h.Write([]byte(rel))
		_, _ = h.Write(b)
		return nil
	})
	if err != nil {
		return out, err
	}
	sortEntries(out.Entries)
	out.EntryCount = len(out.Entries)
	out.SHA256 = hex.EncodeToString(h.Sum(nil))
	return out, nil
}

func snapshotBytes(label, source, filename string, data []byte) (ContentSnapshot, error) {
	sum := sha256.Sum256(data)
	out := ContentSnapshot{Label: label, Source: source, SHA256: hex.EncodeToString(sum[:]), Categories: map[string]int{}}
	low := strings.ToLower(filename)
	if !strings.HasSuffix(low, ".jar") && !strings.HasSuffix(low, ".zip") {
		ent := makeEntry(filename, data)
		out.Entries = []ContentEntry{ent}
		out.EntryCount = 1
		out.Categories[ent.Category] = 1
		return out, nil
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return out, err
	}
	names := make([]string, 0, len(zr.File))
	seen := map[string]bool{}
	for _, f := range zr.File {
		if !f.FileInfo().IsDir() {
			names = append(names, filepath.ToSlash(f.Name))
		}
	}
	prefix := commonArchiveRoot(names)
	if strings.HasSuffix(low, ".jar") {
		prefix = ""
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rawName := filepath.ToSlash(f.Name)
		if unsafeArchivePath(rawName) {
			return out, fmt.Errorf("unsafe archive entry %s", rawName)
		}
		n := normalizeArchivePath(rawName, prefix)
		if n == "" {
			continue
		}
		key := strings.ToLower(n)
		if seen[key] {
			return out, fmt.Errorf("duplicate archive entry %s", n)
		}
		seen[key] = true
		rc, err := f.Open()
		if err != nil {
			return out, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, 256<<20))
		closeErr := rc.Close()
		if err != nil {
			return out, err
		}
		if closeErr != nil {
			return out, closeErr
		}
		ent := makeEntry(n, b)
		out.Entries = append(out.Entries, ent)
		out.Categories[ent.Category]++
	}
	sortEntries(out.Entries)
	out.EntryCount = len(out.Entries)
	return out, nil
}

func commonArchiveRoot(names []string) string {
	if len(names) == 0 {
		return ""
	}
	var root string
	for _, n := range names {
		p := strings.Split(strings.TrimPrefix(n, "/"), "/")
		if len(p) < 2 {
			return ""
		}
		if root == "" {
			root = p[0]
		} else if root != p[0] {
			return ""
		}
	}
	return root + "/"
}
func unsafeArchivePath(n string) bool {
	n = filepath.ToSlash(n)
	clean := filepath.ToSlash(filepath.Clean(n))
	return strings.HasPrefix(n, "/") || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, ":/")
}

func normalizeArchivePath(n, prefix string) string {
	n = strings.TrimPrefix(n, "/")
	if prefix != "" {
		n = strings.TrimPrefix(n, prefix)
	}
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(n)), "./")
}
func makeEntry(path string, b []byte) ContentEntry {
	s := sha256.Sum256(b)
	c := contentCategory(path)
	return ContentEntry{Path: path, Category: c, SHA256: hex.EncodeToString(s[:]), Size: int64(len(b)), Strict: isStrictContent(path, c), Portable: isPortableContent(path, c), Symbols: sourceSymbols(path, b)}
}
func sortEntries(v []ContentEntry) {
	sort.Slice(v, func(i, j int) bool { return v[i].Path < v[j].Path })
}

func contentCategory(path string) string {
	p := strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.HasSuffix(p, ".class"):
		return "code"
	case strings.HasSuffix(p, ".java") || strings.HasSuffix(p, ".kt") || strings.HasSuffix(p, ".scala"):
		return "source"
	case strings.Contains(p, "/textures/"):
		return "textures"
	case strings.Contains(p, "/blockstates/"):
		return "blockstates"
	case strings.Contains(p, "/models/"):
		return "models"
	case strings.Contains(p, "/lang/"):
		return "language"
	case strings.Contains(p, "/sounds/") || strings.HasSuffix(p, "sounds.json"):
		return "sounds"
	case strings.Contains(p, "/recipes/") || strings.Contains(p, "/recipe/"):
		return "recipes"
	case strings.Contains(p, "/loot_tables/") || strings.Contains(p, "/loot_table/"):
		return "loot"
	case strings.Contains(p, "/tags/"):
		return "tags"
	case strings.Contains(p, "/advancements/") || strings.Contains(p, "/advancement/"):
		return "advancements"
	case strings.Contains(p, "/worldgen/"):
		return "worldgen"
	case strings.Contains(p, "/structures/") || strings.Contains(p, "/structure/"):
		return "structures"
	case strings.HasPrefix(p, "assets/"):
		return "assets-other"
	case strings.HasPrefix(p, "data/"):
		return "data-other"
	case p == "fabric.mod.json" || p == "quilt.mod.json" || strings.Contains(p, "mods.toml") || strings.Contains(p, "neoforge.mods.toml"):
		return "metadata"
	case strings.Contains(p, "mixin") && strings.HasSuffix(p, ".json"):
		return "mixins"
	case strings.HasPrefix(p, "meta-inf/"):
		return "metadata"
	default:
		return "other"
	}
}
func isStrictContent(path, cat string) bool {
	p := strings.ToLower(path)
	return strings.HasPrefix(p, "assets/") || strings.HasPrefix(p, "data/") || p == "pack.mcmeta" || cat == "mixins"
}
func isPortableContent(path, cat string) bool {
	switch cat {
	case "textures", "language", "sounds":
		return true
	}
	return false
}

var typeRx = regexp.MustCompile(`(?m)\b(?:class|interface|enum|record|object)\s+([A-Za-z_$][A-Za-z0-9_$]*)`)

func sourceSymbols(path string, b []byte) []string {
	c := contentCategory(path)
	if c != "source" {
		return nil
	}
	m := typeRx.FindAllSubmatch(b, -1)
	seen := map[string]bool{}
	out := []string{}
	for _, x := range m {
		if len(x) > 1 {
			v := string(x[1])
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	sort.Strings(out)
	return out
}
