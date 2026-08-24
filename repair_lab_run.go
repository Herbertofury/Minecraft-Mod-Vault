package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (a *App) startRepairCommand(run *PortingBuildRun, request RepairRunRequest) (*PortingBuildRun, error) {
	if run == nil {
		return nil, errors.New("nil repair session")
	}
	if run.State == "running" {
		return nil, errRepairSessionBusy
	}
	if request.ConfirmCode != repairLabConfirmationPhrase {
		return nil, errors.New("explicit build-script execution acknowledgement is required")
	}
	action := strings.ToLower(strings.TrimSpace(request.Action))
	if action != "build" && action != "test" && action != "clean" {
		return nil, errors.New("action must be build, test, or clean")
	}
	command, workDir, err := repairCommandForRun(run, action)
	if err != nil {
		return nil, err
	}
	if err := verifyImmutableSource(run); err != nil {
		return nil, err
	}
	timeoutMin := clampInt(request.TimeoutMin, 1, 60)
	if request.TimeoutMin <= 0 {
		timeoutMin = 20
	}
	runID := "command-" + time.Now().UTC().Format("20060102T150405.000000000Z")
	logPath := filepath.Join(run.Paths.Logs, runID+".log")
	commandRun := RepairCommandRun{
		ID: runID, Action: action, State: "running", Command: append([]string(nil), command...),
		WorkingDirectory: workDir, StartedAt: time.Now().UTC().Format(time.RFC3339), ExitCode: -1, LogFile: logPath,
	}
	updated, err := a.mutatePortingRun(run.ID, func(current *PortingBuildRun) error {
		if current.State == "running" {
			return errRepairSessionBusy
		}
		current.State = "running"
		current.Phase = "executing-" + action
		current.LastError = ""
		current.Runs = append(current.Runs, commandRun)
		return writeRepairReceipt(current)
	})
	if err != nil {
		return nil, err
	}
	go a.executeRepairCommand(run.ID, runID, action, command, workDir, time.Duration(timeoutMin)*time.Minute)
	return updated, nil
}

func repairCommandForRun(run *PortingBuildRun, action string) ([]string, string, error) {
	if run.Project.Wrapper == "" {
		return nil, "", errors.New("the detected project has no checked-in build wrapper; Repair Lab refuses an unpinned host build tool")
	}
	projectRoot := filepath.Join(run.Paths.WorkingCopy, filepath.FromSlash(run.Project.ProjectRoot))
	wrapper := filepath.Join(projectRoot, filepath.FromSlash(run.Project.Wrapper))
	if !pathContainedBy(projectRoot, wrapper) {
		return nil, "", errors.New("build wrapper escapes the project root")
	}
	info, err := os.Lstat(wrapper)
	if err != nil {
		return nil, "", fmt.Errorf("build wrapper is unavailable: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, "", errors.New("build wrapper must be a regular non-symlink file")
	}
	if run.Project.WrapperSHA256 != "" {
		digest, _, err := hashFileSHA256(wrapper)
		if err != nil {
			return nil, "", err
		}
		if digest != run.Project.WrapperSHA256 {
			return nil, "", errors.New("build wrapper identity changed after import; execution was blocked")
		}
	}
	var args []string
	for _, choice := range run.Project.AvailableCommands {
		if choice.ID == action {
			args = append([]string(nil), choice.Arguments...)
			break
		}
	}
	if len(args) == 0 {
		return nil, "", fmt.Errorf("action %q is unavailable for the detected build system", action)
	}
	if run.Project.BuildSystem == "maven" {
		localRepo := filepath.Join(run.Paths.Root, "tool-cache", "maven")
		args = append([]string{"-Dmaven.repo.local=" + localRepo}, args...)
	}
	if runtime.GOOS == "windows" && (strings.HasSuffix(strings.ToLower(wrapper), ".bat") || strings.HasSuffix(strings.ToLower(wrapper), ".cmd")) {
		command := []string{"cmd.exe", "/d", "/s", "/c", wrapper}
		command = append(command, args...)
		return command, projectRoot, nil
	}
	if err := os.Chmod(wrapper, info.Mode().Perm()|0o100); err != nil {
		return nil, "", err
	}
	return append([]string{wrapper}, args...), projectRoot, nil
}

func (a *App) executeRepairCommand(sessionID, commandID, action string, command []string, workDir string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	a.portingMu.Lock()
	a.portingCancels[sessionID] = cancel
	a.portingMu.Unlock()
	defer func() {
		cancel()
		a.portingMu.Lock()
		delete(a.portingCancels, sessionID)
		a.portingMu.Unlock()
	}()

	run, err := a.loadPortingRun(sessionID)
	if err != nil {
		return
	}
	logPath := filepath.Join(run.Paths.Logs, commandID+".log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		a.finishRepairCommand(sessionID, commandID, action, -1, false, false, err, nil)
		return
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		a.finishRepairCommand(sessionID, commandID, action, -1, false, false, err, nil)
		return
	}
	_, _ = fmt.Fprintf(logFile, "Minecraft Mod Vault Repair Lab\nSession: %s\nStarted: %s\nWorking directory: %s\nCommand: %s\nTimeout: %s\n\n", sessionID, time.Now().UTC().Format(time.RFC3339), workDir, quoteRepairCommand(command), timeout)

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = workDir
	cmd.Env = sanitizedRepairEnvironment(run)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	startErr := cmd.Start()
	if startErr != nil {
		_, _ = fmt.Fprintf(logFile, "\nFailed to start: %v\n", startErr)
		_ = logFile.Close()
		a.finishRepairCommand(sessionID, commandID, action, -1, false, false, startErr, nil)
		return
	}
	waitErr := cmd.Wait()
	_, _ = fmt.Fprintf(logFile, "\nFinished: %s\n", time.Now().UTC().Format(time.RFC3339))
	_ = logFile.Sync()
	_ = logFile.Close()

	exitCode := 0
	if waitErr != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	cancelled := errors.Is(ctx.Err(), context.Canceled) && !timedOut
	artifacts, scanErr := scanRepairArtifacts(run)
	if waitErr == nil && scanErr != nil {
		waitErr = scanErr
		exitCode = -1
	}
	a.finishRepairCommand(sessionID, commandID, action, exitCode, timedOut, cancelled, waitErr, artifacts)
}

func (a *App) finishRepairCommand(sessionID, commandID, action string, exitCode int, timedOut, cancelled bool, runErr error, artifacts []RepairArtifact) {
	_, _ = a.mutatePortingRun(sessionID, func(run *PortingBuildRun) error {
		for i := range run.Runs {
			if run.Runs[i].ID != commandID {
				continue
			}
			run.Runs[i].FinishedAt = time.Now().UTC().Format(time.RFC3339)
			run.Runs[i].ExitCode = exitCode
			run.Runs[i].TimedOut = timedOut
			run.Runs[i].Cancelled = cancelled
			run.Runs[i].LogTail = tailTextFile(run.Runs[i].LogFile, 96<<10)
			switch {
			case timedOut:
				run.Runs[i].State = "timed-out"
				run.Runs[i].Error = "command exceeded its Repair Lab timeout"
			case cancelled:
				run.Runs[i].State = "cancelled"
				run.Runs[i].Error = "command cancelled"
			case runErr != nil:
				run.Runs[i].State = "failed"
				run.Runs[i].Error = runErr.Error()
			default:
				run.Runs[i].State = "succeeded"
			}
			break
		}
		run.Artifacts = artifacts
		if err := verifyImmutableSource(run); err != nil {
			run.Security.SourceVerifiedAfterRuns = false
			run.State = "integrity-failed"
			run.Phase = "immutable-source-check-failed"
			run.LastError = err.Error()
			return writeRepairReceipt(run)
		}
		run.Security.SourceVerifiedAfterRuns = true
		switch {
		case timedOut:
			run.State = "timed-out"
			run.Phase = "execution-timeout"
			run.LastError = "Build execution timed out. The log and working copy were preserved."
		case cancelled:
			run.State = "cancelled"
			run.Phase = "execution-cancelled"
			run.LastError = "Build execution was cancelled. The log and working copy were preserved."
		case runErr != nil:
			run.State = "failed"
			run.Phase = "execution-failed"
			run.LastError = runErr.Error()
		default:
			if action == "clean" {
				run.State = "prepared"
				run.Phase = "clean-succeeded"
			} else {
				run.State = "succeeded"
				run.Phase = action + "-succeeded"
			}
			run.LastError = ""
		}
		return writeRepairReceipt(run)
	})
}

func sanitizedRepairEnvironment(run *PortingBuildRun) []string {
	allow := map[string]bool{
		"PATH": true, "JAVA_HOME": true, "SYSTEMROOT": true, "WINDIR": true, "COMSPEC": true,
		"PATHEXT": true, "LANG": true, "LC_ALL": true, "TERM": true, "USER": true, "USERNAME": true,
	}
	var env []string
	for _, value := range os.Environ() {
		key, _, ok := strings.Cut(value, "=")
		if !ok || !allow[strings.ToUpper(key)] {
			continue
		}
		env = append(env, value)
	}
	home := filepath.Join(run.Paths.Root, "execution-home")
	gradle := filepath.Join(run.Paths.Root, "tool-cache", "gradle")
	temp := filepath.Join(run.Paths.Root, "temp")
	for _, dir := range []string{home, gradle, temp, filepath.Join(run.Paths.Root, "tool-cache", "maven")} {
		_ = os.MkdirAll(dir, 0o755)
	}
	overrides := map[string]string{
		"HOME": home, "USERPROFILE": home, "GRADLE_USER_HOME": gradle, "TMPDIR": temp, "TMP": temp, "TEMP": temp,
		"MMV_REPAIR_LAB": "1", "CI": "true", "GIT_TERMINAL_PROMPT": "0", "MAVEN_OPTS": "-Duser.home=" + home,
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	sort.Strings(env)
	return env
}

func quoteRepairCommand(command []string) string {
	parts := make([]string, 0, len(command))
	for _, value := range command {
		if strings.ContainsAny(value, " \t\"'") {
			parts = append(parts, strconv.Quote(value))
		} else {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, " ")
}

func (a *App) cancelRepairCommand(sessionID string) error {
	a.portingMu.RLock()
	cancel := a.portingCancels[sessionID]
	a.portingMu.RUnlock()
	if cancel == nil {
		return errors.New("no Repair Lab command is running for this session")
	}
	cancel()
	return nil
}

func scanRepairArtifacts(run *PortingBuildRun) ([]RepairArtifact, error) {
	projectRoot := filepath.Join(run.Paths.WorkingCopy, filepath.FromSlash(run.Project.ProjectRoot))
	var paths []string
	err := filepath.WalkDir(projectRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == projectRoot {
			return nil
		}
		if entry.IsDir() {
			name := strings.ToLower(entry.Name())
			if name == ".gradle" || name == ".git" || name == ".idea" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(entry.Name())
		if !strings.HasSuffix(lower, ".jar") && !strings.HasSuffix(lower, ".zip") && !strings.HasSuffix(lower, ".tar.gz") {
			return nil
		}
		rel, _ := filepath.Rel(projectRoot, path)
		rel = filepath.ToSlash(rel)
		if !(strings.Contains(rel, "/build/libs/") || strings.HasPrefix(rel, "build/libs/") || strings.Contains(rel, "/target/") || strings.HasPrefix(rel, "target/") || strings.Contains(rel, "/build/distributions/") || strings.HasPrefix(rel, "build/distributions/")) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	artifacts := make([]RepairArtifact, 0, len(paths))
	for _, path := range paths {
		digest, size, err := hashFileSHA256(path)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(projectRoot, path)
		artifact := RepairArtifact{Name: filepath.Base(path), RelativePath: filepath.ToSlash(rel), Size: size, SHA256: digest, Kind: "archive", DownloadIndex: len(artifacts)}
		if strings.HasSuffix(strings.ToLower(path), ".jar") {
			artifact.Kind = "jar"
			if signals, err := inspectJarSignals(path); err == nil {
				artifact.ModIDs = signals.ModIDs
				artifact.ClassCount = signals.ClassCount
				artifact.JavaMajor = signals.MaxJava
			}
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func createRepairExports(run *PortingBuildRun, includeArtifacts bool) ([]RepairExport, error) {
	if run == nil {
		return nil, errors.New("nil repair session")
	}
	if run.State == "running" {
		return nil, errRepairSessionBusy
	}
	if err := verifyImmutableSource(run); err != nil {
		return nil, err
	}
	if err := writeRepairReceipt(run); err != nil {
		return nil, err
	}
	baseName := safeFilename(strings.TrimSuffix(run.Source.Filename, filepath.Ext(run.Source.Filename)))
	if baseName == "" {
		baseName = "minecraft-mod-source"
	}
	preparedName := fmt.Sprintf("%s-%s-prepared-source.zip", baseName, run.ID)
	preparedPath := filepath.Join(run.Paths.Exports, preparedName)
	digest, size, err := zipDirectoryDeterministic(run.Paths.WorkingCopy, preparedPath, func(rel string, entry fs.DirEntry) bool {
		parts := strings.Split(strings.ToLower(rel), "/")
		for _, part := range parts {
			if part == ".git" || part == ".gradle" || part == "build" || part == "target" || part == "out" || part == "node_modules" || part == ".idea" || part == ".mmv-cache" {
				return false
			}
		}
		return !strings.HasSuffix(strings.ToLower(rel), ".log")
	})
	if err != nil {
		return nil, err
	}
	exports := []RepairExport{{Kind: "prepared-source", Name: preparedName, RelativePath: relativeSessionPath(run.Paths.Root, preparedPath), Size: size, SHA256: digest, CreatedAt: time.Now().UTC().Format(time.RFC3339), DownloadIndex: 0}}

	staging := filepath.Join(run.Paths.Root, ".bundle-staging")
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(filepath.Join(staging, "receipts"), 0o755); err != nil {
		return nil, err
	}
	defer os.RemoveAll(staging)
	if err := copyRepairFilePreserve(preparedPath, filepath.Join(staging, preparedName)); err != nil {
		return nil, err
	}
	for _, source := range []string{run.Paths.ReceiptJSON, run.Paths.ReceiptMarkdown} {
		if pathExists(source) {
			if err := copyRepairFilePreserve(source, filepath.Join(staging, "receipts", filepath.Base(source))); err != nil {
				return nil, err
			}
		}
	}
	if pathExists(run.Paths.Logs) {
		if err := copyDir(run.Paths.Logs, filepath.Join(staging, "logs")); err != nil {
			return nil, err
		}
	}
	if includeArtifacts && len(run.Artifacts) > 0 {
		artifactRoot := filepath.Join(staging, "artifacts")
		if err := os.MkdirAll(artifactRoot, 0o755); err != nil {
			return nil, err
		}
		projectRoot := filepath.Join(run.Paths.WorkingCopy, filepath.FromSlash(run.Project.ProjectRoot))
		for i, artifact := range run.Artifacts {
			source := filepath.Join(projectRoot, filepath.FromSlash(artifact.RelativePath))
			target := filepath.Join(artifactRoot, fmt.Sprintf("%02d-%s", i+1, safeFilename(artifact.Name)))
			if err := copyRepairFilePreserve(source, target); err != nil {
				return nil, err
			}
		}
	}
	bundleName := fmt.Sprintf("Minecraft-Mod-Vault-%s-repair-bundle.zip", run.ID)
	bundlePath := filepath.Join(run.Paths.Exports, bundleName)
	bundleDigest, bundleSize, err := zipDirectoryDeterministic(staging, bundlePath, nil)
	if err != nil {
		return nil, err
	}
	exports = append(exports, RepairExport{Kind: "repair-bundle", Name: bundleName, RelativePath: relativeSessionPath(run.Paths.Root, bundlePath), Size: bundleSize, SHA256: bundleDigest, CreatedAt: time.Now().UTC().Format(time.RFC3339), DownloadIndex: 1})
	return exports, nil
}

func copyRepairFilePreserve(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", source)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
