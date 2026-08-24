package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	testGridSchemaVersion  = 1
	testGridLogMemoryLimit = 8 << 20
	testGridArtifactLimit  = 512 << 20
)

type TestGridManifest struct {
	SchemaVersion     int                `json:"schemaVersion"`
	Name              string             `json:"name"`
	Description       string             `json:"description,omitempty"`
	Edition           string             `json:"edition"`
	GameVersion       string             `json:"gameVersion,omitempty"`
	Loader            string             `json:"loader,omitempty"`
	TimeoutSeconds    int                `json:"timeoutSeconds,omitempty"`
	ExpectedExitCodes []int              `json:"expectedExitCodes,omitempty"`
	Runtime           TestGridRuntime    `json:"runtime"`
	Steps             []TestGridStep     `json:"steps,omitempty"`
	Artifacts         []TestGridArtifact `json:"artifacts,omitempty"`
}

type TestGridRuntime struct {
	Kind             string            `json:"kind"`
	Executable       string            `json:"executable,omitempty"`
	JavaPath         string            `json:"javaPath,omitempty"`
	Jar              string            `json:"jar,omitempty"`
	JVMArguments     []string          `json:"jvmArguments,omitempty"`
	Arguments        []string          `json:"arguments,omitempty"`
	WorkingDirectory string            `json:"workingDirectory,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	Stop             *TestGridStop     `json:"stop,omitempty"`
}

type TestGridStop struct {
	RCONAddress  string `json:"rconAddress,omitempty"`
	PasswordEnv  string `json:"passwordEnv,omitempty"`
	Command      string `json:"command,omitempty"`
	GraceSeconds int    `json:"graceSeconds,omitempty"`
}

type TestGridStep struct {
	Name                 string            `json:"name,omitempty"`
	Type                 string            `json:"type"`
	TimeoutSeconds       int               `json:"timeoutSeconds,omitempty"`
	IntervalMilliseconds int               `json:"intervalMilliseconds,omitempty"`
	Pattern              string            `json:"pattern,omitempty"`
	Address              string            `json:"address,omitempty"`
	Command              string            `json:"command,omitempty"`
	PasswordEnv          string            `json:"passwordEnv,omitempty"`
	Path                 string            `json:"path,omitempty"`
	SHA256               string            `json:"sha256,omitempty"`
	Executable           string            `json:"executable,omitempty"`
	Arguments            []string          `json:"arguments,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
	Optional             bool              `json:"optional,omitempty"`
}

type TestGridArtifact struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path"`
	Optional bool   `json:"optional,omitempty"`
}

type TestGridStepResult struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Status     string         `json:"status"`
	StartedAt  string         `json:"startedAt"`
	FinishedAt string         `json:"finishedAt"`
	DurationMS int64          `json:"durationMs"`
	Message    string         `json:"message,omitempty"`
	Evidence   map[string]any `json:"evidence,omitempty"`
}

type TestGridCapturedArtifact struct {
	Name       string `json:"name"`
	SourcePath string `json:"sourcePath"`
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Size       int64  `json:"size"`
}

type TestGridProcessResult struct {
	PID        int      `json:"pid,omitempty"`
	ExitCode   int      `json:"exitCode"`
	UserCPU    string   `json:"userCpu,omitempty"`
	SystemCPU  string   `json:"systemCpu,omitempty"`
	StoppedBy  string   `json:"stoppedBy,omitempty"`
	Executable string   `json:"executable"`
	Arguments  []string `json:"arguments,omitempty"`
}

type TestGridRun struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	Status       string                     `json:"status"`
	StartedAt    string                     `json:"startedAt"`
	FinishedAt   string                     `json:"finishedAt,omitempty"`
	DurationMS   int64                      `json:"durationMs,omitempty"`
	Edition      string                     `json:"edition"`
	GameVersion  string                     `json:"gameVersion,omitempty"`
	Loader       string                     `json:"loader,omitempty"`
	ManifestPath string                     `json:"manifestPath"`
	RunDirectory string                     `json:"runDirectory"`
	LogPath      string                     `json:"logPath"`
	ReportPath   string                     `json:"reportPath,omitempty"`
	JUnitPath    string                     `json:"junitPath,omitempty"`
	HTMLPath     string                     `json:"htmlPath,omitempty"`
	Error        string                     `json:"error,omitempty"`
	Steps        []TestGridStepResult       `json:"steps,omitempty"`
	Artifacts    []TestGridCapturedArtifact `json:"artifacts,omitempty"`
	Process      TestGridProcessResult      `json:"process"`
}

type testGridRunState struct {
	mu     sync.RWMutex
	run    TestGridRun
	cancel context.CancelFunc
}

type TestGrid struct {
	root string
	mu   sync.RWMutex
	runs map[string]*testGridRunState
}

type synchronizedLog struct {
	mu  sync.RWMutex
	buf []byte
}

func (l *synchronizedLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	l.buf = append(l.buf, p...)
	if len(l.buf) > testGridLogMemoryLimit {
		l.buf = append([]byte(nil), l.buf[len(l.buf)-testGridLogMemoryLimit:]...)
	}
	l.mu.Unlock()
	return len(p), nil
}

func (l *synchronizedLog) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return string(l.buf)
}

func newTestGrid(root string) (*TestGrid, error) {
	root = filepath.Clean(root)
	if root == "." || root == "" {
		return nil, errors.New("TestGrid root is required")
	}
	if err := os.MkdirAll(filepath.Join(root, "runs"), 0o755); err != nil {
		return nil, err
	}
	return &TestGrid{root: root, runs: map[string]*testGridRunState{}}, nil
}

func (g *TestGrid) Start(parent context.Context, manifest TestGridManifest) (TestGridRun, error) {
	manifest = normalizeTestGridManifest(manifest)
	if err := validateTestGridManifest(manifest); err != nil {
		return TestGridRun{}, err
	}
	id := "tg-" + time.Now().UTC().Format("20060102-150405") + "-" + randomToken(4)
	runDir := filepath.Join(g.root, "runs", id)
	if err := os.MkdirAll(filepath.Join(runDir, "artifacts"), 0o755); err != nil {
		return TestGridRun{}, err
	}
	started := time.Now().UTC()
	run := TestGridRun{
		ID: id, Name: manifest.Name, Status: "queued", StartedAt: started.Format(time.RFC3339Nano),
		Edition: manifest.Edition, GameVersion: manifest.GameVersion, Loader: manifest.Loader,
		ManifestPath: filepath.Join(runDir, "manifest.json"), RunDirectory: runDir,
		LogPath: filepath.Join(runDir, "process.log"),
		Process: TestGridProcessResult{ExitCode: -1},
	}
	ctx, cancel := context.WithCancel(parent)
	state := &testGridRunState{run: run, cancel: cancel}
	g.mu.Lock()
	g.runs[id] = state
	g.mu.Unlock()
	if err := writeTestGridJSONAtomic(run.ManifestPath, redactTestGridManifest(manifest)); err != nil {
		cancel()
		return TestGridRun{}, err
	}
	go g.execute(ctx, state, manifest, started)
	return g.snapshot(state), nil
}

func (g *TestGrid) Run(parent context.Context, manifest TestGridManifest) (TestGridRun, error) {
	run, err := g.Start(parent, manifest)
	if err != nil {
		return TestGridRun{}, err
	}
	for {
		current, ok := g.Get(run.ID)
		if !ok {
			return TestGridRun{}, errors.New("TestGrid run disappeared")
		}
		switch current.Status {
		case "passed":
			return current, nil
		case "failed", "canceled":
			if current.Error == "" {
				current.Error = "TestGrid run " + current.Status
			}
			return current, errors.New(current.Error)
		}
		select {
		case <-parent.Done():
			_ = g.Cancel(run.ID)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (g *TestGrid) Get(id string) (TestGridRun, bool) {
	g.mu.RLock()
	state, ok := g.runs[id]
	g.mu.RUnlock()
	if ok {
		return g.snapshot(state), true
	}
	var run TestGridRun
	if err := readTestGridJSON(filepath.Join(g.root, "runs", filepath.Base(id), "report.json"), &run); err != nil {
		return TestGridRun{}, false
	}
	return run, true
}

func (g *TestGrid) List() []TestGridRun {
	g.mu.RLock()
	states := make([]*testGridRunState, 0, len(g.runs))
	seen := make(map[string]bool, len(g.runs))
	for _, state := range g.runs {
		states = append(states, state)
	}
	g.mu.RUnlock()
	runs := make([]TestGridRun, 0, len(states))
	for _, state := range states {
		run := g.snapshot(state)
		runs = append(runs, run)
		seen[run.ID] = true
	}
	entries, _ := os.ReadDir(filepath.Join(g.root, "runs"))
	for _, entry := range entries {
		if !entry.IsDir() || seen[entry.Name()] {
			continue
		}
		var run TestGridRun
		if readTestGridJSON(filepath.Join(g.root, "runs", entry.Name(), "report.json"), &run) == nil {
			runs = append(runs, run)
		}
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt > runs[j].StartedAt })
	return runs
}

func (g *TestGrid) Cancel(id string) bool {
	g.mu.RLock()
	state, ok := g.runs[id]
	g.mu.RUnlock()
	if !ok {
		return false
	}
	state.cancel()
	return true
}

func (g *TestGrid) CancelAll() {
	g.mu.RLock()
	states := make([]*testGridRunState, 0, len(g.runs))
	for _, state := range g.runs {
		states = append(states, state)
	}
	g.mu.RUnlock()
	for _, state := range states {
		state.cancel()
	}
}

func (g *TestGrid) snapshot(state *testGridRunState) TestGridRun {
	state.mu.RLock()
	defer state.mu.RUnlock()
	run := state.run
	run.Steps = append([]TestGridStepResult(nil), state.run.Steps...)
	run.Artifacts = append([]TestGridCapturedArtifact(nil), state.run.Artifacts...)
	run.Process.Arguments = append([]string(nil), state.run.Process.Arguments...)
	return run
}

func (g *TestGrid) execute(parent context.Context, state *testGridRunState, manifest TestGridManifest, started time.Time) {
	timeout := time.Duration(manifest.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	setRun := func(fn func(*TestGridRun)) {
		state.mu.Lock()
		fn(&state.run)
		state.mu.Unlock()
	}
	setRun(func(run *TestGridRun) { run.Status = "running" })

	executable, args, workDir, env, err := resolveTestGridCommand(manifest.Runtime)
	if err != nil {
		g.finish(state, started, "failed", err)
		return
	}
	setRun(func(run *TestGridRun) {
		run.Process.Executable = executable
		run.Process.Arguments = append([]string(nil), args...)
	})
	logFile, err := os.OpenFile(state.run.LogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		g.finish(state, started, "failed", err)
		return
	}
	defer logFile.Close()
	memoryLog := &synchronizedLog{}
	logWriter := io.MultiWriter(logFile, memoryLog)
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err := cmd.Start(); err != nil {
		g.finish(state, started, "failed", fmt.Errorf("start runtime: %w", err))
		return
	}
	setRun(func(run *TestGridRun) { run.Process.PID = cmd.Process.Pid })
	waitDone := make(chan struct{})
	var waitErr error
	go func() {
		waitErr = cmd.Wait()
		close(waitDone)
	}()

	var firstErr error
	for index, step := range manifest.Steps {
		result := g.executeStep(ctx, step, workDir, manifest.Runtime.Environment, memoryLog, waitDone)
		if result.Name == "" {
			result.Name = fmt.Sprintf("step-%02d", index+1)
		}
		setRun(func(run *TestGridRun) { run.Steps = append(run.Steps, result) })
		if result.Status == "failed" && !step.Optional {
			firstErr = errors.New(result.Message)
			break
		}
	}

	stoppedBy := ""
	select {
	case <-waitDone:
	default:
		stoppedBy = g.stopRuntime(ctx, manifest.Runtime, cmd, waitDone)
	}
	if stoppedBy != "" {
		setRun(func(run *TestGridRun) { run.Process.StoppedBy = stoppedBy })
	}
	select {
	case <-waitDone:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		<-waitDone
	}
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
		setRun(func(run *TestGridRun) {
			run.Process.ExitCode = exitCode
			run.Process.UserCPU = cmd.ProcessState.UserTime().String()
			run.Process.SystemCPU = cmd.ProcessState.SystemTime().String()
		})
	}
	if firstErr == nil && stoppedBy == "" && !containsInt(manifest.ExpectedExitCodes, exitCode) {
		if waitErr == nil {
			firstErr = fmt.Errorf("runtime exited with code %d; expected %v", exitCode, manifest.ExpectedExitCodes)
		} else {
			firstErr = fmt.Errorf("runtime exited with code %d: %w", exitCode, waitErr)
		}
	}

	artifacts, artifactErr := captureTestGridArtifacts(manifest.Artifacts, workDir, filepath.Join(state.run.RunDirectory, "artifacts"))
	setRun(func(run *TestGridRun) { run.Artifacts = artifacts })
	if firstErr == nil && artifactErr != nil {
		firstErr = artifactErr
	}
	if firstErr == nil && ctx.Err() != nil {
		firstErr = ctx.Err()
	}
	if firstErr != nil {
		status := "failed"
		if errors.Is(firstErr, context.Canceled) {
			status = "canceled"
		}
		g.finish(state, started, status, firstErr)
		return
	}
	g.finish(state, started, "passed", nil)
}

func (g *TestGrid) executeStep(ctx context.Context, step TestGridStep, workDir string, runtimeEnv map[string]string, logBuffer *synchronizedLog, processDone <-chan struct{}) TestGridStepResult {
	started := time.Now().UTC()
	name := strings.TrimSpace(step.Name)
	if name == "" {
		name = strings.TrimSpace(step.Type)
	}
	result := TestGridStepResult{Name: name, Type: step.Type, Status: "passed", StartedAt: started.Format(time.RFC3339Nano)}
	stepTimeout := time.Duration(step.TimeoutSeconds) * time.Second
	stepCtx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	var evidence map[string]any
	var err error
	switch step.Type {
	case "wait-log":
		err = waitForLog(stepCtx, logBuffer, step.Pattern, step.IntervalMilliseconds, processDone)
	case "assert-log":
		err = assertLog(logBuffer.String(), step.Pattern, true)
	case "deny-log":
		err = assertLog(logBuffer.String(), step.Pattern, false)
	case "tcp":
		err = waitForTCP(stepCtx, expandTestGridValue(step.Address), step.IntervalMilliseconds)
	case "java-ping":
		var status map[string]any
		status, err = waitForJavaStatus(stepCtx, expandTestGridValue(step.Address), step.IntervalMilliseconds)
		evidence = status
	case "bedrock-ping":
		var status BedrockStatus
		status, err = waitForBedrockStatus(stepCtx, expandTestGridValue(step.Address), step.IntervalMilliseconds)
		if err == nil {
			body, _ := json.Marshal(status)
			_ = json.Unmarshal(body, &evidence)
		}
	case "rcon":
		password, passwordErr := testGridSecret(step.PasswordEnv, runtimeEnv)
		if passwordErr != nil {
			err = passwordErr
			break
		}
		var response string
		response, err = minecraftRCON(stepCtx, expandTestGridValue(step.Address), password, step.Command)
		if response != "" {
			evidence = map[string]any{"response": response}
		}
	case "file-exists":
		path := resolveTestGridPath(workDir, step.Path)
		var info os.FileInfo
		info, err = waitForTestGridFile(stepCtx, path, step.IntervalMilliseconds)
		if err == nil {
			evidence = map[string]any{"path": path, "size": info.Size(), "directory": info.IsDir()}
		}
	case "sha256":
		path := resolveTestGridPath(workDir, step.Path)
		var sum string
		sum, _, err = hashTestGridFile(path)
		if err == nil && !strings.EqualFold(sum, strings.TrimSpace(step.SHA256)) {
			err = fmt.Errorf("SHA-256 mismatch for %s: got %s", path, sum)
		}
		if sum != "" {
			evidence = map[string]any{"path": path, "sha256": sum}
		}
	case "command":
		executable := expandTestGridValue(step.Executable)
		if executable == "" {
			err = errors.New("command step requires executable")
			break
		}
		args := expandTestGridValues(step.Arguments)
		command := exec.CommandContext(stepCtx, executable, args...)
		command.Dir = workDir
		command.Env = mergeTestGridEnvironment(runtimeEnv, step.Environment)
		var output []byte
		output, err = command.CombinedOutput()
		message := strings.TrimSpace(string(output))
		if len(message) > 4096 {
			message = message[len(message)-4096:]
		}
		evidence = map[string]any{"output": message}
	case "sleep":
		select {
		case <-stepCtx.Done():
			err = stepCtx.Err()
		case <-time.After(time.Duration(step.IntervalMilliseconds) * time.Millisecond):
		}
	default:
		err = fmt.Errorf("unsupported TestGrid step type %q", step.Type)
	}
	if err != nil {
		if step.Optional {
			result.Status = "skipped"
			result.Message = err.Error()
		} else {
			result.Status = "failed"
			result.Message = err.Error()
		}
	}
	finished := time.Now().UTC()
	result.FinishedAt = finished.Format(time.RFC3339Nano)
	result.DurationMS = finished.Sub(started).Milliseconds()
	result.Evidence = evidence
	return result
}

func (g *TestGrid) stopRuntime(ctx context.Context, runtimeSpec TestGridRuntime, cmd *exec.Cmd, done <-chan struct{}) string {
	grace := 10 * time.Second
	if runtimeSpec.Stop != nil && runtimeSpec.Stop.GraceSeconds > 0 {
		grace = time.Duration(runtimeSpec.Stop.GraceSeconds) * time.Second
	}
	if runtimeSpec.Stop != nil && runtimeSpec.Stop.RCONAddress != "" && runtimeSpec.Stop.Command != "" {
		password, err := testGridSecret(runtimeSpec.Stop.PasswordEnv, runtimeSpec.Environment)
		if err == nil {
			stopCtx, cancel := context.WithTimeout(context.Background(), grace)
			_, rconErr := minecraftRCON(stopCtx, expandTestGridValue(runtimeSpec.Stop.RCONAddress), password, runtimeSpec.Stop.Command)
			cancel()
			if rconErr == nil {
				select {
				case <-done:
					return "rcon"
				case <-time.After(grace):
				}
			}
		}
	}
	if cmd.Process != nil {
		_ = cmd.Process.Signal(os.Interrupt)
		select {
		case <-done:
			return "interrupt"
		case <-time.After(grace):
		}
		_ = cmd.Process.Kill()
		return "kill"
	}
	_ = ctx
	return ""
}

func (g *TestGrid) finish(state *testGridRunState, started time.Time, status string, runErr error) {
	finished := time.Now().UTC()
	state.mu.Lock()
	state.run.FinishedAt = finished.Format(time.RFC3339Nano)
	state.run.DurationMS = finished.Sub(started).Milliseconds()
	state.run.ReportPath = filepath.Join(state.run.RunDirectory, "report.json")
	state.run.JUnitPath = filepath.Join(state.run.RunDirectory, "junit.xml")
	state.run.HTMLPath = filepath.Join(state.run.RunDirectory, "report.html")
	run := state.run
	run.Status = status
	if runErr != nil {
		run.Error = runErr.Error()
	}
	state.mu.Unlock()
	writeErrors := make([]string, 0, 3)
	if err := writeTestGridJSONAtomic(run.ReportPath, run); err != nil {
		writeErrors = append(writeErrors, "JSON report: "+err.Error())
	}
	if err := writeTestGridJUnit(run.JUnitPath, run); err != nil {
		writeErrors = append(writeErrors, "JUnit report: "+err.Error())
	}
	if err := writeTestGridHTML(run.HTMLPath, run); err != nil {
		writeErrors = append(writeErrors, "HTML report: "+err.Error())
	}
	if len(writeErrors) > 0 {
		status = "failed"
		if run.Error != "" {
			run.Error += "; "
		}
		run.Error += strings.Join(writeErrors, "; ")
	}
	state.mu.Lock()
	state.run.Status = status
	state.run.Error = run.Error
	state.mu.Unlock()
}

func normalizeTestGridManifest(manifest TestGridManifest) TestGridManifest {
	if manifest.SchemaVersion == 0 {
		manifest.SchemaVersion = testGridSchemaVersion
	}
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Edition = strings.ToLower(strings.TrimSpace(manifest.Edition))
	manifest.Loader = strings.ToLower(strings.TrimSpace(manifest.Loader))
	manifest.Runtime.Kind = strings.ToLower(strings.TrimSpace(manifest.Runtime.Kind))
	if manifest.TimeoutSeconds <= 0 {
		manifest.TimeoutSeconds = 300
	}
	if len(manifest.ExpectedExitCodes) == 0 {
		manifest.ExpectedExitCodes = []int{0}
	}
	for index := range manifest.Steps {
		manifest.Steps[index].Type = strings.ToLower(strings.TrimSpace(manifest.Steps[index].Type))
		if manifest.Steps[index].TimeoutSeconds <= 0 {
			manifest.Steps[index].TimeoutSeconds = 30
		}
		if manifest.Steps[index].IntervalMilliseconds <= 0 {
			manifest.Steps[index].IntervalMilliseconds = 100
		}
	}
	return manifest
}

func validateTestGridManifest(manifest TestGridManifest) error {
	if manifest.SchemaVersion != testGridSchemaVersion {
		return fmt.Errorf("unsupported TestGrid schemaVersion %d; expected %d", manifest.SchemaVersion, testGridSchemaVersion)
	}
	if manifest.Name == "" {
		return errors.New("TestGrid manifest name is required")
	}
	switch manifest.Edition {
	case "java", "bedrock", "cross-edition", "generic":
	default:
		return fmt.Errorf("unsupported edition %q", manifest.Edition)
	}
	if manifest.TimeoutSeconds < 1 || manifest.TimeoutSeconds > 86400 {
		return errors.New("timeoutSeconds must be between 1 and 86400")
	}
	if _, _, _, _, err := resolveTestGridCommand(manifest.Runtime); err != nil {
		return err
	}
	supported := map[string]bool{"wait-log": true, "assert-log": true, "deny-log": true, "tcp": true, "java-ping": true, "bedrock-ping": true, "rcon": true, "file-exists": true, "sha256": true, "command": true, "sleep": true}
	for index, step := range manifest.Steps {
		if !supported[step.Type] {
			return fmt.Errorf("step %d has unsupported type %q", index+1, step.Type)
		}
		if step.TimeoutSeconds < 1 || step.TimeoutSeconds > manifest.TimeoutSeconds {
			return fmt.Errorf("step %d timeoutSeconds must be between 1 and the run timeout", index+1)
		}
		if (step.Type == "wait-log" || step.Type == "assert-log" || step.Type == "deny-log") && step.Pattern == "" {
			return fmt.Errorf("step %d requires pattern", index+1)
		}
		if step.Pattern != "" {
			if _, err := regexp.Compile(step.Pattern); err != nil {
				return fmt.Errorf("step %d pattern: %w", index+1, err)
			}
		}
	}
	return nil
}

func resolveTestGridCommand(runtimeSpec TestGridRuntime) (string, []string, string, []string, error) {
	workDir := expandTestGridValue(runtimeSpec.WorkingDirectory)
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", nil, "", nil, err
		}
	}
	workDir, _ = filepath.Abs(filepath.Clean(workDir))
	if info, err := os.Stat(workDir); err != nil || !info.IsDir() {
		return "", nil, "", nil, fmt.Errorf("runtime working directory is unavailable: %s", workDir)
	}
	executable := expandTestGridValue(runtimeSpec.Executable)
	args := expandTestGridValues(runtimeSpec.Arguments)
	if runtimeSpec.Jar != "" {
		if executable == "" {
			executable = expandTestGridValue(runtimeSpec.JavaPath)
		}
		if executable == "" {
			executable = "java"
		}
		jar := resolveTestGridPath(workDir, runtimeSpec.Jar)
		args = append(append(expandTestGridValues(runtimeSpec.JVMArguments), "-jar", jar), args...)
	}
	if executable == "" {
		return "", nil, "", nil, errors.New("runtime executable or Java jar is required")
	}
	return executable, args, workDir, mergeTestGridEnvironment(runtimeSpec.Environment, nil), nil
}

func redactTestGridManifest(manifest TestGridManifest) TestGridManifest {
	manifest.Runtime.Environment = redactEnvironment(manifest.Runtime.Environment)
	for index := range manifest.Steps {
		manifest.Steps[index].Environment = redactEnvironment(manifest.Steps[index].Environment)
	}
	return manifest
}

func redactEnvironment(environment map[string]string) map[string]string {
	if environment == nil {
		return nil
	}
	result := make(map[string]string, len(environment))
	for key, value := range environment {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "key") {
			value = "[redacted]"
		}
		result[key] = value
	}
	return result
}

func mergeTestGridEnvironment(primary, secondary map[string]string) []string {
	values := map[string]string{}
	for _, item := range os.Environ() {
		if index := strings.IndexByte(item, '='); index > 0 {
			values[item[:index]] = item[index+1:]
		}
	}
	for key, value := range primary {
		values[key] = expandTestGridValue(value)
	}
	for key, value := range secondary {
		values[key] = expandTestGridValue(value)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}

func testGridSecret(name string, environment map[string]string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("passwordEnv is required; TestGrid never stores an RCON password in a manifest")
	}
	if value, ok := environment[name]; ok && value != "" {
		return expandTestGridValue(value), nil
	}
	if value := os.Getenv(name); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("environment variable %s is empty", name)
}

func expandTestGridValue(value string) string { return os.ExpandEnv(strings.TrimSpace(value)) }

func expandTestGridValues(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = os.ExpandEnv(value)
	}
	return result
}

func resolveTestGridPath(workDir, path string) string {
	path = expandTestGridValue(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	return filepath.Clean(path)
}

func waitForTestGridFile(ctx context.Context, path string, intervalMS int) (os.FileInfo, error) {
	interval := time.Duration(intervalMS) * time.Millisecond
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		info, err := os.Stat(path)
		if err == nil {
			return info, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("file %s did not appear: %w", path, ctx.Err())
		case <-ticker.C:
		}
	}
}

func waitForLog(ctx context.Context, logBuffer *synchronizedLog, pattern string, intervalMS int, processDone <-chan struct{}) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(intervalMS) * time.Millisecond)
	defer ticker.Stop()
	for {
		if re.MatchString(logBuffer.String()) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("log pattern %q not observed: %w", pattern, ctx.Err())
		case <-processDone:
			if re.MatchString(logBuffer.String()) {
				return nil
			}
			return fmt.Errorf("runtime exited before log pattern %q appeared", pattern)
		case <-ticker.C:
		}
	}
}

func assertLog(logText, pattern string, wantMatch bool) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	matched := re.MatchString(logText)
	if wantMatch && !matched {
		return fmt.Errorf("log pattern %q was not found", pattern)
	}
	if !wantMatch && matched {
		return fmt.Errorf("forbidden log pattern %q was found", pattern)
	}
	return nil
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func captureTestGridArtifacts(specs []TestGridArtifact, workDir, artifactDir string) ([]TestGridCapturedArtifact, error) {
	result := make([]TestGridCapturedArtifact, 0, len(specs))
	for index, spec := range specs {
		source := resolveTestGridPath(workDir, spec.Path)
		info, err := os.Lstat(source)
		if err != nil {
			if spec.Optional {
				continue
			}
			return result, fmt.Errorf("artifact %s: %w", spec.Path, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return result, fmt.Errorf("artifact %s is a symbolic link", spec.Path)
		}
		name := sanitizeTestGridArtifactName(spec.Name)
		if name == "" {
			name = sanitizeTestGridArtifactName(filepath.Base(source))
		}
		if name == "" {
			name = fmt.Sprintf("artifact-%02d", index+1)
		}
		destination := filepath.Join(artifactDir, name)
		if info.IsDir() {
			destination += ".zip"
			err = zipTestGridDirectory(source, destination)
		} else {
			err = copyTestGridArtifact(source, destination, info.Size())
		}
		if err != nil {
			return result, fmt.Errorf("capture artifact %s: %w", spec.Path, err)
		}
		sum, size, err := hashTestGridFile(destination)
		if err != nil {
			return result, err
		}
		result = append(result, TestGridCapturedArtifact{Name: name, SourcePath: source, Path: destination, SHA256: sum, Size: size})
	}
	return result, nil
}

func sanitizeTestGridArtifactName(value string) string {
	value = strings.TrimSpace(value)
	value = regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(value, "-")
	return strings.Trim(value, ".-")
}

func copyTestGridArtifact(source, destination string, size int64) error {
	if size > testGridArtifactLimit {
		return fmt.Errorf("file exceeds %d-byte artifact limit", testGridArtifactLimit)
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(out, io.LimitReader(in, testGridArtifactLimit+1))
	closeErr := out.Close()
	if written > testGridArtifactLimit {
		return errors.New("artifact exceeded size limit while copying")
	}
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func zipTestGridDirectory(source, destination string) error {
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	var total int64
	walkErr := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		if total > testGridArtifactLimit {
			return errors.New("directory exceeds artifact size limit")
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relative)
		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, input)
		closeErr := input.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	zipCloseErr := zw.Close()
	fileCloseErr := out.Close()
	if walkErr != nil {
		return walkErr
	}
	if zipCloseErr != nil {
		return zipCloseErr
	}
	return fileCloseErr
}

func hashTestGridFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func writeTestGridJSONAtomic(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".tmp-" + randomToken(3)
	if err := os.WriteFile(temporary, body, 0o644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

func readTestGridJSON(path string, value any) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func testGridCapabilities() map[string]any {
	return map[string]any{
		"schemaVersion": testGridSchemaVersion,
		"host":          map[string]string{"os": runtime.GOOS, "arch": runtime.GOARCH},
		"runtimes": []map[string]any{
			{"id": "java-server", "headless": true, "versions": "all versions supported by the supplied Java and server distribution", "evidence": []string{"process", "logs", "TCP", "server-list ping", "RCON", "files", "hashes"}},
			{"id": "bedrock-server", "headless": true, "versions": "all versions supported by the supplied Bedrock Dedicated Server", "evidence": []string{"process", "logs", "TCP", "RakNet status ping", "files", "hashes"}},
			{"id": "build-or-tool", "headless": true, "versions": "tool-defined", "evidence": []string{"process", "logs", "exit code", "files", "hashes"}},
			{"id": "java-client-render", "headless": "driver-dependent", "versions": "loader and render-driver dependent", "evidence": []string{"screenshots or video supplied as declared artifacts"}},
			{"id": "bedrock-retail-client", "headless": false, "versions": "Windows retail client is not a supported headless server process", "evidence": []string{"use a real-client automation/render adapter and declare its captures as artifacts"}},
		},
		"steps":    []string{"wait-log", "assert-log", "deny-log", "tcp", "java-ping", "bedrock-ping", "rcon", "file-exists", "sha256", "command", "sleep"},
		"reports":  []string{"json", "junit-xml", "html", "combined-process-log", "sha256-addressed-artifacts"},
		"security": []string{"loopback-token API", "no shell interpolation", "RCON secrets referenced by environment variable", "redacted persisted environment", "bounded artifact capture"},
	}
}
