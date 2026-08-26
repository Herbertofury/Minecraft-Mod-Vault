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

const version = "2.0.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version":
		fmt.Println("Minecraft Dev Kit Orchestrator", version)
	case "validate":
		runValidate(os.Args[2:])
	case "check":
		runCheck(os.Args[2:])
	case "sync":
		runSync(os.Args[2:])
	case "watch":
		runWatch(os.Args[2:])
	case "sources":
		runSources(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}
func usage() {
	fmt.Print(`Minecraft Dev Kit Orchestrator 2.0.0

Commands:
  mmv-devkit validate --registry devkit-registry.json
  mmv-devkit check --registry devkit-registry.json [--mc 1.21.1] [--loader neoforge] [--json]
  mmv-devkit sync --registry devkit-registry.json [--mc ...] [--loader ...] [--apply] [--drive]
  mmv-devkit watch --registry devkit-registry.json --drive [--interval 15m]
  mmv-devkit sources --catalog tool-provenance.json [--id geckolib-forge] [--json]

Credentials are read from environment only:
  CURSEFORGE_API_KEY, GITHUB_TOKEN, MMV_GOOGLE_DRIVE_TOKEN
`)
}
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

type provenanceCatalog struct {
	Schema int              `json:"schema"`
	Count  int              `json:"count"`
	Tools  []provenanceTool `json:"tools"`
}
type provenanceTool struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Homepage           string        `json:"homepage"`
	Patterns           []string      `json:"patterns"`
	UpdateProviders    []ProviderRef `json:"updateProviders"`
	AutoUpdateEligible bool          `json:"autoUpdateEligible"`
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
