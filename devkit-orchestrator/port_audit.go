package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func entryMap(s ContentSnapshot) map[string]ContentEntry {
	m := map[string]ContentEntry{}
	for _, e := range s.Entries {
		m[strings.ToLower(e.Path)] = e
	}
	return m
}
func diffSnapshots(a, b ContentSnapshot) ContentDelta {
	am, bm := entryMap(a), entryMap(b)
	d := ContentDelta{}
	for k, x := range bm {
		if y, ok := am[k]; !ok {
			d.Added = append(d.Added, x)
		} else if y.SHA256 != x.SHA256 {
			d.Changed = append(d.Changed, ContentChange{Path: x.Path, Category: x.Category, BeforeHash: y.SHA256, AfterHash: x.SHA256})
		}
	}
	for k, x := range am {
		if _, ok := bm[k]; !ok {
			d.Removed = append(d.Removed, x)
		}
	}
	sortEntries(d.Added)
	sortEntries(d.Removed)
	sort.Slice(d.Changed, func(i, j int) bool { return d.Changed[i].Path < d.Changed[j].Path })
	return d
}
func depKey(d Dependency) string { return strings.ToLower(d.Provider + ":" + d.ProjectID) }
func diffDependencies(a, b []Dependency) DependencyDelta {
	am := map[string]Dependency{}
	bm := map[string]Dependency{}
	for _, d := range a {
		if d.Required {
			am[depKey(d)] = d
		}
	}
	for _, d := range b {
		if d.Required {
			bm[depKey(d)] = d
		}
	}
	out := DependencyDelta{}
	for k, d := range bm {
		if _, ok := am[k]; !ok {
			out.AddedRequired = append(out.AddedRequired, d)
		}
	}
	for k, d := range am {
		if _, ok := bm[k]; !ok {
			out.RemovedRequired = append(out.RemovedRequired, d)
		}
	}
	return out
}

func auditPort(original, converted ContentSnapshot, convertedSource *ContentSnapshot, h HeritageReport) []AuditFinding {
	om, cm, lm, tm := entryMap(original), entryMap(converted), entryMap(h.RuntimeLatest), entryMap(h.RuntimeCompatible)
	findings := []AuditFinding{}
	add := func(sev, kind, path, cat, msg, exp, act string) {
		findings = append(findings, AuditFinding{Severity: sev, Kind: kind, Path: path, Category: cat, Message: msg, Expected: exp, Actual: act})
	}
	for k, o := range om {
		if !o.Strict {
			continue
		}
		l, lok := lm[k]
		if !lok {
			add("info", "upstream-removed", o.Path, o.Category, "original content was removed by newest upstream and is not mandatory in the conversion", "", "")
			continue
		}
		c, cok := cm[k]
		if !cok {
			add("error", "original-content-missing", o.Path, o.Category, "content present in the original and still present upstream is missing from the converted target", l.SHA256, "")
			continue
		}
		if o.SHA256 == l.SHA256 && c.SHA256 != o.SHA256 {
			add("error", "unchanged-content-corrupted", o.Path, o.Category, "upstream did not change this original content, but the converted target did", o.SHA256, c.SHA256)
			continue
		}
		if o.SHA256 != l.SHA256 {
			if c.SHA256 == o.SHA256 {
				add("error", "stale-original-content", o.Path, o.Category, "new upstream content exists but the converted target still carries the original bytes", l.SHA256, c.SHA256)
			} else if l.Portable && c.SHA256 != l.SHA256 {
				add("error", "latest-portable-content-mismatch", o.Path, o.Category, "portable upstream content should be copied exactly from the newest release", l.SHA256, c.SHA256)
			} else if c.SHA256 != l.SHA256 {
				add("warning", "adapted-latest-content", o.Path, o.Category, "latest content differs from the conversion; verify this is a required target-version adaptation", l.SHA256, c.SHA256)
			}
		}
	}
	for k, l := range lm {
		if !l.Strict {
			continue
		}
		if _, old := om[k]; old {
			continue
		}
		c, ok := cm[k]
		if !ok {
			add("error", "latest-feature-missing", l.Path, l.Category, "new content exists in the newest upstream release but is absent from the converted target", l.SHA256, "")
			continue
		}
		if l.Portable && c.SHA256 != l.SHA256 {
			add("error", "latest-feature-corrupted", l.Path, l.Category, "new portable content does not match newest upstream bytes", l.SHA256, c.SHA256)
		} else if c.SHA256 != l.SHA256 {
			add("warning", "latest-feature-adapted", l.Path, l.Category, "new upstream content is present but adapted; verify target-version compatibility", l.SHA256, c.SHA256)
		}
	}
	for k, t := range tm {
		if !t.Strict {
			continue
		}
		if l, ok := lm[k]; ok && t.SHA256 != l.SHA256 {
			if c, ok := cm[k]; ok && c.SHA256 == t.SHA256 {
				add("error", "target-version-stale-content", c.Path, c.Category, "converted target matches the older target-compatible release while newer upstream content exists", l.SHA256, c.SHA256)
			}
		}
	}
	if convertedSource != nil && h.SourceCompatible != nil && h.SourceLatest != nil {
		before := entryMap(*h.SourceCompatible)
		latest := entryMap(*h.SourceLatest)
		symbols := map[string]bool{}
		for _, e := range convertedSource.Entries {
			for _, s := range e.Symbols {
				symbols[s] = true
			}
		}
		for k, e := range latest {
			if e.Category != "source" {
				continue
			}
			if _, old := before[k]; old {
				continue
			}
			for _, s := range e.Symbols {
				if !symbols[s] {
					add("error", "latest-source-surface-missing", e.Path, "source", "new upstream source type "+s+" is not represented in the converted source tree", "", "")
				}
			}
		}
	} else if h.SourceDelta != nil && len(h.SourceDelta.Added) > 0 {
		add("error", "converted-source-not-supplied", "", "source", "new upstream source files exist; pass --converted-source so the completion gate can prove their type surface was carried into the conversion", "", "")
	}
	for _, d := range h.DependencyDelta.AddedRequired {
		add("warning", "latest-required-dependency-added", "", "dependency", "new upstream required dependency must be evaluated/backported: "+d.Provider+":"+d.ProjectID, "", "")
	}
	sort.SliceStable(findings, func(i, j int) bool {
		rank := func(s string) int {
			if s == "error" {
				return 0
			}
			if s == "warning" {
				return 1
			}
			return 2
		}
		ri, rj := rank(findings[i].Severity), rank(findings[j].Severity)
		if ri != rj {
			return ri < rj
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

func writeReport(path string, v any) error {
	if path == "" {
		return nil
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
