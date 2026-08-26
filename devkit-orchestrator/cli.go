package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

func loadRegistry(path string) (Registry, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return Registry{}, e
	}
	var r Registry
	e = json.Unmarshal(b, &r)
	if e != nil {
		return r, e
	}
	return r, validateRegistry(r)
}
func common(fs *flag.FlagSet) (*string, *string, *string, *string, *string) {
	reg := fs.String("registry", "devkit-registry.json", "registry path")
	mc := fs.String("mc", "", "Minecraft override")
	loader := fs.String("loader", "", "loader override")
	channel := fs.String("channel", "", "release channel override")
	arch := fs.String("arch", runtime.GOARCH, "architecture override")
	return reg, mc, loader, channel, arch
}
func mkTarget(mc, loader, channel, arch string) Target {
	return Target{Minecraft: mc, Loader: loader, Channel: channel, OS: runtime.GOOS, Arch: arch}
}
func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	p := fs.String("registry", "devkit-registry.json", "registry path")
	_ = fs.Parse(args)
	r, e := loadRegistry(*p)
	fatal(e)
	fmt.Printf("ok: %d artifacts, schema %d\n", len(r.Artifacts), r.Schema)
}
func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	reg, mc, loader, channel, arch := common(fs)
	js := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)
	r, e := loadRegistry(*reg)
	fatal(e)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	p, e := newEngine(r, nil).plan(ctx, mkTarget(*mc, *loader, *channel, *arch))
	fatal(e)
	if *js {
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(b))
		return
	}
	for _, a := range p.Artifacts {
		fmt.Printf("%-12s %-38s %s\n", a.Status, a.ArtifactID, a.Reason)
	}
}
func runSync(args []string) {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	reg, mc, loader, channel, arch := common(fs)
	apply := fs.Bool("apply", false, "apply updates")
	drive := fs.Bool("drive", false, "update Drive file IDs in place")
	lock := fs.String("lock", "devkit-lock.json", "lockfile path")
	_ = fs.Parse(args)
	r, e := loadRegistry(*reg)
	fatal(e)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	eng := newEngine(r, nil)
	p, e := eng.plan(ctx, mkTarget(*mc, *loader, *channel, *arch))
	fatal(e)
	if !*apply {
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(b))
		return
	}
	if !*drive {
		fatal(fmt.Errorf("--apply requires --drive so verified bytes have a durable destination"))
	}
	l, e := eng.apply(ctx, p, *reg, *lock, *drive)
	fatal(e)
	fmt.Printf("updated %d artifacts; lockfile %s\n", len(l.Entries), *lock)
}
func fatal(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, "error:", e)
		os.Exit(1)
	}
}

func runWatch(args []string) {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	reg, mc, loader, channel, arch := common(fs)
	drive := fs.Bool("drive", false, "update Drive in place")
	interval := fs.Duration("interval", 15*time.Minute, "refresh interval")
	lock := fs.String("lock", "devkit-lock.json", "lockfile path")
	_ = fs.Parse(args)
	if !*drive {
		fatal(fmt.Errorf("watch requires --drive"))
	}
	if *interval < time.Minute {
		fatal(fmt.Errorf("watch interval must be at least 1m"))
	}
	for {
		r, e := loadRegistry(*reg)
		if e != nil {
			fmt.Fprintln(os.Stderr, "watch:", e)
			time.Sleep(*interval)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		eng := newEngine(r, nil)
		p, e := eng.plan(ctx, mkTarget(*mc, *loader, *channel, *arch))
		if e == nil {
			_, e = eng.apply(ctx, p, *reg, *lock, true)
		}
		cancel()
		if e != nil {
			fmt.Fprintln(os.Stderr, "watch:", e)
		} else {
			fmt.Println(time.Now().Format(time.RFC3339), "refresh complete")
		}
		time.Sleep(*interval)
	}
}

func runSources(args []string) {
	fs := flag.NewFlagSet("sources", flag.ExitOnError)
	catalogPath := fs.String("catalog", "tool-provenance.json", "tool provenance catalog")
	id := fs.String("id", "", "optional exact tool id")
	asJSON := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)
	b, err := os.ReadFile(*catalogPath)
	fatal(err)
	var c provenanceCatalog
	fatal(json.Unmarshal(b, &c))
	selected := []provenanceTool{}
	for _, t := range c.Tools {
		if *id == "" || strings.EqualFold(t.ID, *id) {
			selected = append(selected, t)
		}
	}
	if *id != "" && len(selected) == 0 {
		fatal(fmt.Errorf("unknown catalog id %s", *id))
	}
	if *asJSON {
		out, _ := json.MarshalIndent(selected, "", "  ")
		fmt.Println(string(out))
		return
	}
	for _, t := range selected {
		provider := "manual"
		if len(t.UpdateProviders) > 0 {
			p := t.UpdateProviders[0]
			provider = p.Type + ":" + first(p.Repo, p.Project, p.URL)
		}
		fmt.Printf("%-28s %-42s %-38s %s\n", t.ID, t.Name, provider, t.Homepage)
	}
}

func runAdoptTools(args []string) {
	fs := flag.NewFlagSet("adopt-tools", flag.ExitOnError)
	regPath := fs.String("registry", "devkit-registry.json", "registry path")
	catalogPath := fs.String("catalog", "tool-provenance.json", "tool provenance catalog")
	drive := fs.Bool("drive", false, "discover existing files from the configured Google Drive root")
	_ = fs.Parse(args)
	if !*drive {
		fatal(fmt.Errorf("adopt-tools requires --drive"))
	}
	reg, err := loadRegistry(*regPath)
	fatal(err)
	cat, err := loadProvenanceCatalog(*catalogPath)
	fatal(err)
	dc, err := newDriveClient(reg.Drive, nil)
	fatal(err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	n, skipped, err := adoptTools(ctx, &reg, cat, dc)
	fatal(err)
	fatal(validateRegistry(reg))
	fatal(writeJSONAtomic(*regPath, reg))
	fmt.Printf("adopted %d existing Drive tool artifacts into live managed state\n", n)
	if len(skipped) > 0 {
		fmt.Printf("skipped %d catalog/file cases; use sources --json to inspect provenance\n", len(skipped))
	}
}

func runHeritage(args []string) {
	fs := flag.NewFlagSet("heritage", flag.ExitOnError)
	reg, mc, loader, channel, arch := common(fs)
	id := fs.String("id", "", "managed mod/library artifact id")
	outDir := fs.String("out", "", "directory for newest + target-compatible runtime/source references")
	report := fs.String("report", "heritage-report.json", "machine-readable report path")
	asJSON := fs.Bool("json", false, "print full JSON report")
	_ = fs.Parse(args)
	if strings.TrimSpace(*id) == "" {
		fatal(fmt.Errorf("heritage requires --id"))
	}
	r, err := loadRegistry(*reg)
	fatal(err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	h, err := newEngine(r, nil).buildHeritage(ctx, *id, mkTarget(*mc, *loader, *channel, *arch), *outDir)
	fatal(err)
	fatal(writeReport(*report, h))
	if *asJSON {
		b, _ := json.MarshalIndent(h, "", "  ")
		fmt.Println(string(b))
		return
	}
	fmt.Printf("feature authority: %s (%s)\n", h.LatestUpstream.Version, h.LatestUpstream.Filename)
	fmt.Printf("target-compatible: %s (%s)\n", h.TargetCompatible.Version, h.TargetCompatible.Filename)
	fmt.Printf("delta: +%d changed=%d removed=%d; releases=%d; report=%s\n", len(h.RuntimeDelta.Added), len(h.RuntimeDelta.Changed), len(h.RuntimeDelta.Removed), len(h.ReleaseLineage), *report)
}

func runPortGuard(args []string) {
	fs := flag.NewFlagSet("port-guard", flag.ExitOnError)
	reg, mc, loader, channel, arch := common(fs)
	id := fs.String("id", "", "managed source mod/library artifact id")
	original := fs.String("original", "", "exact original mod artifact being converted/worked on")
	converted := fs.String("converted", "", "converted target build/JAR to audit")
	convertedSource := fs.String("converted-source", "", "converted project source directory or source ZIP")
	outDir := fs.String("out", "", "directory for upstream heritage references")
	report := fs.String("report", "port-guard-report.json", "machine-readable completion-gate report")
	asJSON := fs.Bool("json", false, "print full JSON report")
	_ = fs.Parse(args)
	if strings.TrimSpace(*id) == "" {
		fatal(fmt.Errorf("port-guard requires --id"))
	}
	r, err := loadRegistry(*reg)
	fatal(err)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	g, err := newEngine(r, nil).portGuard(ctx, *id, mkTarget(*mc, *loader, *channel, *arch), *original, *converted, *convertedSource, *outDir)
	fatal(err)
	fatal(writeReport(*report, g))
	if *asJSON {
		b, _ := json.MarshalIndent(g, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("port guard: passed=%v errors=%d warnings=%d report=%s\n", g.Passed, g.Errors, g.Warnings, *report)
		for _, f := range g.Findings {
			if f.Severity == "info" {
				continue
			}
			fmt.Printf("%-7s %-32s %s %s\n", f.Severity, f.Kind, f.Path, f.Message)
		}
	}
	if !g.Passed {
		os.Exit(3)
	}
}
