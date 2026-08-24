package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const conversionAdapterTimeout = 30 * time.Minute

func (a *App) handleConversionToolConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request ConversionToolConfigRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	request.ToolID = strings.TrimSpace(request.ToolID)
	known := false
	for _, tool := range conversionToolCatalog() {
		if tool.ID == request.ToolID {
			known = true
			break
		}
	}
	if !known {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "unknown conversion tool"})
		return
	}
	path := strings.TrimSpace(request.Path)
	if path != "" {
		absolute, err := filepath.Abs(path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		info, err := os.Lstat(absolute)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			writeJSON(w, http.StatusBadRequest, APIError{Error: "tool path must be an existing regular file, not a symbolic link"})
			return
		}
		path = filepath.Clean(absolute)
	}
	a.mu.Lock()
	if a.settings.ConversionToolPaths == nil {
		a.settings.ConversionToolPaths = map[string]string{}
	}
	if path == "" {
		delete(a.settings.ConversionToolPaths, request.ToolID)
	} else {
		a.settings.ConversionToolPaths[request.ToolID] = path
	}
	a.mu.Unlock()
	if err := a.saveSettings(); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	for _, tool := range a.configuredConversionToolAdapters() {
		if tool.ID == request.ToolID {
			writeJSON(w, http.StatusOK, tool)
			return
		}
	}
	writeJSON(w, http.StatusInternalServerError, APIError{Error: "configured tool disappeared from catalog"})
}

func (a *App) handleConversionAdapterRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request ConversionAdapterRunRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	session, err := a.loadConversionSession(request.SessionID)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	if session.Plan == nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "build a conversion plan before running a specialist adapter"})
		return
	}
	var tool ConversionToolAdapter
	found := false
	for _, candidate := range a.configuredConversionToolAdapters() {
		if candidate.ID == strings.TrimSpace(request.ToolID) {
			tool, found = candidate, true
			break
		}
	}
	if !found || !tool.Ready || !tool.CanExecute {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "the selected adapter is not configured for allowlisted execution"})
		return
	}
	run, err := a.executeConversionAdapter(r.Context(), session, tool, request.Options)
	if saveErr := a.saveConversionSession(session); saveErr != nil && err == nil {
		err = saveErr
	}
	if err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session, "adapterRun": run})
}

func (a *App) executeConversionAdapter(parent context.Context, session *ConversionSession, tool ConversionToolAdapter, options map[string]string) (*ConversionAdapterRun, error) {
	if err := verifyConversionSource(session); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	run := ConversionAdapterRun{ID: fmt.Sprintf("adapter-%s-%s", now.Format("20060102T150405Z"), randomToken(5)), ToolID: tool.ID, ToolName: tool.Name, State: "running", StartedAt: now.Format(time.RFC3339)}
	session.AdapterRuns = append(session.AdapterRuns, run)
	runIndex := len(session.AdapterRuns) - 1
	root := filepath.Join(session.Paths.Workspace, "adapters", run.ID)
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		return nil, err
	}
	logPath := filepath.Join(root, "adapter.log")
	command, outputPaths, err := buildConversionAdapterCommand(session, tool, work, options)
	if err != nil {
		run.State, run.Error, run.CompletedAt = "failed", err.Error(), time.Now().UTC().Format(time.RFC3339)
		session.AdapterRuns[runIndex] = run
		return &session.AdapterRuns[runIndex], err
	}
	run.WorkingDir, run.LogPath = work, logPath
	run.Command = append([]string{command.Path}, command.Args[1:]...)
	ctx, cancel := context.WithTimeout(parent, conversionAdapterTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command.Path, command.Args[1:]...)
	cmd.Dir = work
	if strings.TrimSpace(command.Dir) != "" {
		cmd.Dir = command.Dir
	}
	cmd.Env = conversionAdapterEnvironment(root)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	_, _ = fmt.Fprintf(logFile, "Minecraft Mod Vault OmniBridge adapter run\nTool: %s (%s)\nStarted: %s\nCommand: %s\n\n", tool.Name, tool.ID, run.StartedAt, strings.Join(run.Command, " "))
	cmd.Stdout, cmd.Stderr = logFile, logFile
	runErr := cmd.Run()
	_ = logFile.Close()
	if ctx.Err() == context.DeadlineExceeded {
		runErr = fmt.Errorf("adapter exceeded %s timeout", conversionAdapterTimeout)
	}
	if exitErr := new(exec.ExitError); errors.As(runErr, &exitErr) {
		run.ExitCode = exitErr.ExitCode()
	}
	if digest, _, hashErr := hashFileSHA256(logPath); hashErr == nil {
		run.LogSHA256 = digest
	}
	if verifyErr := verifyConversionSource(session); verifyErr != nil {
		runErr = fmt.Errorf("adapter changed immutable source: %w", verifyErr)
	} else {
		run.SourceVerified = true
	}
	if runErr == nil {
		for _, outputPath := range outputPaths() {
			if !pathContainedBy(root, outputPath) || !regularFile(outputPath) {
				continue
			}
			final := filepath.Join(session.Paths.Outputs, filepath.Base(outputPath))
			if err := copyFileReplace(outputPath, final); err != nil {
				runErr = err
				break
			}
			record, err := conversionOutputRecord(final, len(session.Outputs))
			if err != nil {
				runErr = err
				break
			}
			record.Kind = "adapter-" + tool.ID
			record.Validated, record.Validation = validateConversionOutput(final, session.Plan.Target.Format)
			session.Outputs = append(session.Outputs, record)
			run.Outputs = append(run.Outputs, record)
		}
		if len(run.Outputs) == 0 && runErr == nil {
			runErr = errors.New("adapter exited successfully but produced no recognized target artifact")
		}
	}
	run.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	if runErr != nil {
		run.State, run.Error = "failed", runErr.Error()
		session.State, session.Phase, session.LastError = "review-required", "adapter-failed", runErr.Error()
	} else {
		run.State = "succeeded"
		session.State, session.Phase, session.LastError = "review-required", "adapter-validated", ""
	}
	session.AdapterRuns[runIndex] = run
	return &session.AdapterRuns[runIndex], runErr
}

type conversionAdapterCommand struct {
	Path string
	Args []string
	Dir  string
}

type conversionAdapterOutputResolver func() []string

func buildConversionAdapterCommand(session *ConversionSession, tool ConversionToolAdapter, work string, options map[string]string) (conversionAdapterCommand, conversionAdapterOutputResolver, error) {
	switch tool.ID {
	case "chunker":
		return buildChunkerCommand(session, tool, work, options)
	case "je2be-resource":
		return buildJE2BEResourceCommand(session, tool, work, options)
	case "packconverter":
		return buildPackConverterCommand(session, tool, work)
	case "regolith":
		return buildRegolithCommand(session, tool, work)
	default:
		return conversionAdapterCommand{}, nil, fmt.Errorf("tool %q has no reviewed execution contract", tool.ID)
	}
}

func buildChunkerCommand(session *ConversionSession, tool ConversionToolAdapter, work string, options map[string]string) (conversionAdapterCommand, conversionAdapterOutputResolver, error) {
	if session.Plan.Target.Format != "bedrock-world" && session.Plan.Target.Format != "bedrock-template" && session.Plan.Target.Format != "java-world" {
		return conversionAdapterCommand{}, nil, errors.New("Chunker can only execute for a world target")
	}
	worldRoot := findConversionWorldRoot(session.Paths.Extracted)
	if worldRoot == "" {
		return conversionAdapterCommand{}, nil, errors.New("Chunker input world could not be located")
	}
	outputDir := filepath.Join(work, "converted-world")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	format := chunkerFormat(session.Plan.Target.Edition, session.Plan.Target.GameVersion)
	args := []string{"-i", worldRoot, "-f", format, "-o", outputDir}
	for option, flag := range map[string]string{"blockMappings": "-m", "worldSettings": "-s", "pruning": "-p", "converterSettings": "-c", "dimensionRegistry": "-r", "dimensionMappings": "-d", "biomeMappings": "-b"} {
		if value := strings.TrimSpace(options[option]); value != "" {
			absolute, err := filepath.Abs(value)
			if err != nil || !regularFile(absolute) {
				return conversionAdapterCommand{}, nil, fmt.Errorf("%s must name an existing settings file", option)
			}
			args = append(args, flag, absolute)
		}
	}
	command, err := executableCommand(tool.DetectedPath, args)
	if err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	resolver := func() []string {
		if !directoryHasFiles(outputDir) {
			return nil
		}
		stage := outputDir
		if session.Plan.Target.Format == "bedrock-template" {
			_ = writeBedrockWorldTemplateMetadata(stage, session)
		}
		name := cleanConversionName(firstNonEmpty(session.Plan.Target.Name, session.Name))
		ext := ".zip"
		if session.Plan.Target.Format == "bedrock-world" {
			ext = ".mcworld"
		} else if session.Plan.Target.Format == "bedrock-template" {
			ext = ".mctemplate"
		}
		path := filepath.Join(work, name+"-chunker"+ext)
		_, _, _ = zipDirectoryDeterministic(stage, path, nil)
		return []string{path}
	}
	return command, resolver, nil
}

func buildJE2BEResourceCommand(session *ConversionSession, tool ConversionToolAdapter, work string, options map[string]string) (conversionAdapterCommand, conversionAdapterOutputResolver, error) {
	if session.Plan.Target.Format != "bedrock-resource" && session.Plan.Target.Format != "bedrock-addon" {
		return conversionAdapterCommand{}, nil, errors.New("JE2BE Resource Pack Converter requires a Bedrock resource/add-on target")
	}
	input, err := prepareJavaResourcePackInput(session, work)
	if err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	output := filepath.Join(work, cleanConversionName(firstNonEmpty(session.Plan.Target.Name, session.Name))+"-je2be.mcpack")
	args := []string{"convert", input, output, "--pack-name", firstNonEmpty(session.Plan.Target.Name, session.Name), "--pack-description", firstNonEmpty(session.Plan.Target.Description, "Converted by Minecraft Mod Vault OmniBridge")}
	if strings.EqualFold(options["essentials"], "true") {
		args = append(args, "--essentials")
	}
	if strings.EqualFold(options["rtxfix"], "true") {
		args = append(args, "--rtxfix")
	}
	if strings.EqualFold(options["disablePBR"], "true") {
		args = append(args, "--disable-pbr")
	}
	command, err := executableCommand(tool.DetectedPath, args)
	if err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	return command, func() []string { return []string{output} }, nil
}

func buildPackConverterCommand(session *ConversionSession, tool ConversionToolAdapter, work string) (conversionAdapterCommand, conversionAdapterOutputResolver, error) {
	if session.Plan.Target.Format != "bedrock-resource" && session.Plan.Target.Format != "bedrock-addon" {
		return conversionAdapterCommand{}, nil, errors.New("PackConverter requires a Bedrock resource/add-on target")
	}
	input, err := prepareJavaResourcePackInput(session, work)
	if err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	command, err := executableCommand(tool.DetectedPath, []string{"nogui", "--input", input})
	if err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	return command, func() []string {
		candidates := []string{}
		_ = filepath.WalkDir(work, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Clean(path) == filepath.Clean(input) {
				return err
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".mcpack" || ext == ".zip" {
				candidates = append(candidates, path)
			}
			return nil
		})
		sort.Strings(candidates)
		return candidates
	}, nil
}

func buildRegolithCommand(session *ConversionSession, tool ConversionToolAdapter, work string) (conversionAdapterCommand, conversionAdapterOutputResolver, error) {
	supported := map[string]bool{
		"bedrock-addon": true, "bedrock-behavior": true, "bedrock-resource": true,
		"bedrock-project": true, "bedrock-world-product": true,
	}
	if !supported[session.Plan.Target.Format] {
		return conversionAdapterCommand{}, nil, errors.New("Regolith requires a Bedrock add-on or project target")
	}
	project := filepath.Join(work, "project")
	if err := buildBedrockProjectDirectory(session, project); err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	command, err := executableCommand(tool.DetectedPath, []string{"run", "default"})
	if err != nil {
		return conversionAdapterCommand{}, nil, err
	}
	command.Dir = project
	exportRoot := filepath.Join(project, "build", "export")
	resolver := func() []string {
		bp := filepath.Join(exportRoot, "BP")
		rp := filepath.Join(exportRoot, "RP")
		if !directoryHasFiles(bp) && !directoryHasFiles(rp) {
			return nil
		}
		stage := filepath.Join(work, "regolith-package")
		_ = os.RemoveAll(stage)
		_ = os.MkdirAll(stage, 0o755)
		if directoryHasFiles(bp) {
			_ = copyDir(bp, filepath.Join(stage, shortPackFolder(session.Plan.Target.Namespace, "BP")))
		}
		if directoryHasFiles(rp) {
			_ = copyDir(rp, filepath.Join(stage, shortPackFolder(session.Plan.Target.Namespace, "RP")))
		}
		output := filepath.Join(work, cleanConversionName(firstNonEmpty(session.Plan.Target.Name, session.Name))+"-regolith.mcaddon")
		if _, _, err := zipDirectoryDeterministic(stage, output, nil); err != nil {
			return nil
		}
		return []string{output}
	}
	return command, resolver, nil
}

func executableCommand(path string, args []string) (conversionAdapterCommand, error) {
	path = filepath.Clean(path)
	if !regularFile(path) {
		return conversionAdapterCommand{}, errors.New("configured adapter executable is unavailable")
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jar":
		java, err := exec.LookPath("java")
		if err != nil {
			return conversionAdapterCommand{}, errors.New("Java is required to run this adapter")
		}
		return conversionAdapterCommand{Path: java, Args: append([]string{java, "-jar", path}, args...)}, nil
	case ".py":
		python, err := exec.LookPath("python3")
		if err != nil {
			python, err = exec.LookPath("python")
		}
		if err != nil {
			return conversionAdapterCommand{}, errors.New("Python is required to run this adapter")
		}
		return conversionAdapterCommand{Path: python, Args: append([]string{python, path}, args...)}, nil
	default:
		return conversionAdapterCommand{Path: path, Args: append([]string{path}, args...)}, nil
	}
}

func prepareJavaResourcePackInput(session *ConversionSession, work string) (string, error) {
	stage := filepath.Join(work, "java-resource-input")
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return "", err
	}
	copied := 0
	for _, candidate := range []string{"assets", "pack.mcmeta", "pack.png"} {
		source := filepath.Join(session.Paths.Extracted, candidate)
		if info, err := os.Stat(source); err == nil {
			if info.IsDir() {
				if err := copyDir(source, filepath.Join(stage, candidate)); err != nil {
					return "", err
				}
			} else if err := copyFileReplace(source, filepath.Join(stage, candidate)); err != nil {
				return "", err
			}
			copied++
		}
	}
	if copied == 0 {
		// JARs and nested pack archives may carry a common top-level folder.
		_ = filepath.WalkDir(session.Paths.Extracted, func(path string, entry os.DirEntry, err error) error {
			if err != nil || !entry.IsDir() || !strings.EqualFold(entry.Name(), "assets") {
				return err
			}
			if copyErr := copyDir(path, filepath.Join(stage, "assets")); copyErr == nil {
				copied++
			}
			return filepath.SkipDir
		})
	}
	if copied == 0 {
		return "", errors.New("no Java resource-pack assets were found for the selected adapter")
	}
	if !pathExists(filepath.Join(stage, "pack.mcmeta")) {
		_ = writeJSONFileAtomic(filepath.Join(stage, "pack.mcmeta"), map[string]any{"pack": map[string]any{"pack_format": javaPackFormat(firstNonEmpty(session.Source.GameVersion, "1.21.1"), "resource"), "description": "OmniBridge adapter input"}})
	}
	input := filepath.Join(work, "java-resource-input.zip")
	if _, _, err := zipDirectoryDeterministic(stage, input, nil); err != nil {
		return "", err
	}
	return input, nil
}

func chunkerFormat(edition, version string) string {
	edition = strings.ToUpper(strings.TrimSpace(edition))
	if edition != "JAVA" && edition != "BEDROCK" {
		edition = "JAVA"
	}
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	version = strings.NewReplacer(".", "_", "-", "_").Replace(version)
	return edition + "_" + strings.ToUpper(version)
}

func conversionAdapterEnvironment(root string) []string {
	allowed := map[string]bool{"PATH": true, "HOME": true, "USERPROFILE": true, "SYSTEMROOT": true, "WINDIR": true, "JAVA_HOME": true, "LANG": true, "LC_ALL": true, "TMPDIR": true, "TEMP": true, "TMP": true}
	env := []string{}
	for _, pair := range os.Environ() {
		key := strings.SplitN(pair, "=", 2)[0]
		if allowed[strings.ToUpper(key)] {
			env = append(env, pair)
		}
	}
	cache := filepath.Join(root, "cache")
	_ = os.MkdirAll(cache, 0o755)
	env = append(env, "GRADLE_USER_HOME="+filepath.Join(cache, "gradle"), "MAVEN_OPTS=-Dmaven.repo.local="+filepath.Join(cache, "maven"), "PYTHONPYCACHEPREFIX="+filepath.Join(cache, "pycache"))
	if runtime.GOOS != "windows" {
		env = append(env, "HOME="+filepath.Join(root, "home"))
		_ = os.MkdirAll(filepath.Join(root, "home"), 0o755)
	}
	return env
}

func regularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func directoryHasFiles(path string) bool {
	found := false
	_ = filepath.WalkDir(path, func(_ string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			found = true
			return io.EOF
		}
		return err
	})
	return found
}
