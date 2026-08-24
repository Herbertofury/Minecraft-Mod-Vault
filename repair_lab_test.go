package main

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type repairZipEntry struct {
	name string
	body string
	mode os.FileMode
}

func writeRepairZip(t *testing.T, path string, entries []repairZipEntry) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		out, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(out, entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func fabricRepairFixture() []repairZipEntry {
	return []repairZipEntry{
		{name: "demo/settings.gradle", body: "rootProject.name = 'repair-demo'\n"},
		{name: "demo/build.gradle", body: "plugins { id 'fabric-loom' version '1.6-SNAPSHOT' }\njava { toolchain.languageVersion = JavaLanguageVersion.of(17) }\n"},
		{name: "demo/gradle.properties", body: "minecraft_version=1.20.1\nloader_version=0.15.11\nloom_version=1.6.12\nyarn_mappings=1.20.1+build.10\njava_version=17\n"},
		{name: "demo/local.properties", body: "secret_local_path=C:/private/do-not-touch\nminecraft_version=1.20.1\n"},
		{name: "demo/gradlew", body: "#!/bin/sh\nexit 0\n", mode: 0o755},
		{name: "demo/src/main/resources/fabric.mod.json", body: "{\n  \"schemaVersion\": 1,\n  \"id\": \"repair_demo\",\n  \"version\": \"1.0.0\",\n  \"depends\": {\n    \"fabricloader\": \">=0.15.11\",\n    \"minecraft\": \"~1.20.1\",\n    \"java\": \">=17\"\n  }\n}\n"},
		{name: "demo/src/main/resources/pack.mcmeta", body: "{\"pack\":{\"pack_format\":15,\"description\":\"demo\"}}\n"},
		{name: "demo/src/main/java/example/Demo.java", body: "package example; public final class Demo {}\n"},
	}
}

func TestSafeExtractSourceZipAcceptsRegularProject(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "source.zip")
	writeRepairZip(t, archive, fabricRepairFixture())
	destination := filepath.Join(t.TempDir(), "source")
	result, err := safeExtractSourceZip(archive, destination)
	if err != nil {
		t.Fatal(err)
	}
	if result.FileCount != len(fabricRepairFixture()) || result.RootHint != "demo" {
		t.Fatalf("unexpected extraction result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(destination, "demo", "gradlew")); err != nil {
		t.Fatal(err)
	}
}

func TestSafeExtractSourceZipRejectsTraversalSymlinkDuplicateAndDevicePaths(t *testing.T) {
	tests := []struct {
		name    string
		entries []repairZipEntry
		want    string
	}{
		{name: "traversal", entries: []repairZipEntry{{name: "../evil.txt", body: "evil"}}, want: "parent traversal"},
		{name: "symlink", entries: []repairZipEntry{{name: "link", body: "target", mode: os.ModeSymlink | 0o777}}, want: "unsupported link"},
		{name: "duplicate", entries: []repairZipEntry{{name: "same.txt", body: "one"}, {name: "same.txt", body: "two"}}, want: "duplicate path"},
		{name: "device", entries: []repairZipEntry{{name: "CON.txt", body: "bad"}}, want: "Windows device path"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			archive := filepath.Join(t.TempDir(), "source.zip")
			writeRepairZip(t, archive, tc.entries)
			_, err := safeExtractSourceZip(archive, filepath.Join(t.TempDir(), "source"))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestRepairProjectDetectionAndPreparedMigrationPreserveImmutableSource(t *testing.T) {
	root := t.TempDir()
	app := &App{cfgDir: root, portingRuns: map[string]*PortingBuildRun{}, portingCancels: map[string]context.CancelFunc{}}
	run, err := app.newPortingRun("Fabric repair fixture")
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "source.zip")
	writeRepairZip(t, archive, fabricRepairFixture())
	if err := copyRepairFilePreserve(archive, run.Paths.OriginalArchive); err != nil {
		t.Fatal(err)
	}
	archiveHash, archiveSize, err := hashFileSHA256(run.Paths.OriginalArchive)
	if err != nil {
		t.Fatal(err)
	}
	extracted, err := safeExtractSourceZip(run.Paths.OriginalArchive, run.Paths.ImmutableSource)
	if err != nil {
		t.Fatal(err)
	}
	treeHash, files, bytes, err := hashDirectoryTree(run.Paths.ImmutableSource)
	if err != nil {
		t.Fatal(err)
	}
	run.Source = RepairSourceSnapshot{Filename: "source.zip", Size: archiveSize, SHA256: archiveHash, TreeSHA256: treeHash, FileCount: files, ExtractedBytes: bytes, ImportedAt: time.Now().UTC().Format(time.RFC3339)}
	if files != extracted.FileCount || bytes != extracted.ExtractedBytes {
		t.Fatalf("source identity mismatch: files=%d/%d bytes=%d/%d", files, extracted.FileCount, bytes, extracted.ExtractedBytes)
	}
	if err := resetWorkingCopy(run); err != nil {
		t.Fatal(err)
	}
	profile, err := detectRepairProject(run.Paths.ImmutableSource)
	if err != nil {
		t.Fatal(err)
	}
	if profile.ProjectRoot != "demo" || profile.BuildSystem != "gradle" || profile.Wrapper != "gradlew" || profile.Loader != "fabric" || profile.GameVersion != "1.20.1" || profile.JavaMajor != 17 {
		t.Fatalf("unexpected detected profile: %#v", profile)
	}
	run.Project = profile
	run.State = "imported"

	if err := prepareRepairSession(run, RepairPrepareRequest{TargetGameVersion: "1.21.1", TargetLoader: "fabric"}); err != nil {
		t.Fatal(err)
	}
	if run.State != "prepared" || run.Target == nil || run.Target.GameVersion != "1.21.1" || len(run.Changes) == 0 {
		t.Fatalf("migration not staged: %#v", run)
	}
	if err := verifyImmutableSource(run); err != nil {
		t.Fatalf("immutable source changed: %v", err)
	}
	immutableProps, err := os.ReadFile(filepath.Join(run.Paths.ImmutableSource, "demo", "gradle.properties"))
	if err != nil {
		t.Fatal(err)
	}
	workingProps, err := os.ReadFile(filepath.Join(run.Paths.WorkingCopy, "demo", "gradle.properties"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(immutableProps), "minecraft_version=1.20.1") || !strings.Contains(string(workingProps), "minecraft_version=1.21.1") {
		t.Fatalf("source/work split failed: immutable=%q working=%q", immutableProps, workingProps)
	}
	immutableLocal, _ := os.ReadFile(filepath.Join(run.Paths.ImmutableSource, "demo", "local.properties"))
	workingLocal, _ := os.ReadFile(filepath.Join(run.Paths.WorkingCopy, "demo", "local.properties"))
	if string(immutableLocal) != string(workingLocal) {
		t.Fatalf("local.properties was rewritten despite being host-local/sensitive: immutable=%q working=%q", immutableLocal, workingLocal)
	}
	fabricMeta, err := os.ReadFile(filepath.Join(run.Paths.WorkingCopy, "demo", "src", "main", "resources", "fabric.mod.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fabricMeta), `"minecraft": "=1.21.1"`) {
		t.Fatalf("Fabric metadata target was not updated: %s", fabricMeta)
	}
}

func TestRepairLabExecutesOnlyAcknowledgedWrapperAndCapturesArtifacts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("synthetic shell-wrapper fixture is exercised on Unix builds")
	}
	wrapper := "#!/bin/sh\nset -eu\nmkdir -p build/libs\nprintf 'synthetic repaired artifact' > build/libs/repair-demo.jar\necho 'repair fixture build completed'\n"
	app, run := setupRepairFixtureRun(t, wrapper)
	if err := prepareRepairSession(run, RepairPrepareRequest{TargetGameVersion: "1.21.1", TargetLoader: "fabric"}); err != nil {
		t.Fatal(err)
	}
	if err := app.savePortingRun(run); err != nil {
		t.Fatal(err)
	}
	if _, err := app.startRepairCommand(run, RepairRunRequest{SessionID: run.ID, Action: "build", ConfirmCode: "wrong", TimeoutMin: 1}); err == nil {
		t.Fatal("build execution succeeded without the exact acknowledgement phrase")
	}
	started, err := app.startRepairCommand(run, RepairRunRequest{SessionID: run.ID, Action: "build", ConfirmCode: repairLabConfirmationPhrase, TimeoutMin: 1})
	if err != nil {
		t.Fatal(err)
	}
	if started.State != "running" {
		t.Fatalf("command did not enter running state: %#v", started)
	}
	deadline := time.Now().Add(15 * time.Second)
	var finished *PortingBuildRun
	for time.Now().Before(deadline) {
		finished, err = app.loadPortingRun(run.ID)
		if err != nil {
			t.Fatal(err)
		}
		if finished.State != "running" {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}
	if finished == nil || finished.State != "succeeded" {
		t.Fatalf("wrapper execution did not succeed: %#v", finished)
	}
	if len(finished.Runs) != 1 || finished.Runs[0].ExitCode != 0 || !strings.Contains(finished.Runs[0].LogTail, "repair fixture build completed") {
		t.Fatalf("command proof incomplete: %#v", finished.Runs)
	}
	if len(finished.Artifacts) != 1 || finished.Artifacts[0].Name != "repair-demo.jar" || finished.Artifacts[0].SHA256 == "" {
		t.Fatalf("artifact proof incomplete: %#v", finished.Artifacts)
	}
	if err := verifyImmutableSource(finished); err != nil {
		t.Fatalf("wrapper altered immutable source: %v", err)
	}
	exports, err := createRepairExports(finished, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(exports) != 2 || exports[0].SHA256 == "" || exports[1].SHA256 == "" {
		t.Fatalf("proof exports incomplete: %#v", exports)
	}
	for _, export := range exports {
		path := filepath.Join(finished.Paths.Root, filepath.FromSlash(export.RelativePath))
		if digest, size, err := hashFileSHA256(path); err != nil || digest != export.SHA256 || size != export.Size {
			t.Fatalf("export verification failed for %s: digest=%s size=%d err=%v", export.Name, digest, size, err)
		}
	}
}

func setupRepairFixtureRun(t *testing.T, wrapper string) (*App, *PortingBuildRun) {
	t.Helper()
	root := t.TempDir()
	app := &App{cfgDir: root, portingRuns: map[string]*PortingBuildRun{}, portingCancels: map[string]context.CancelFunc{}}
	run, err := app.newPortingRun("Executable Fabric repair fixture")
	if err != nil {
		t.Fatal(err)
	}
	entries := fabricRepairFixture()
	for i := range entries {
		if entries[i].name == "demo/gradlew" {
			entries[i].body = wrapper
			entries[i].mode = 0o755
		}
	}
	archive := filepath.Join(t.TempDir(), "source.zip")
	writeRepairZip(t, archive, entries)
	if err := copyRepairFilePreserve(archive, run.Paths.OriginalArchive); err != nil {
		t.Fatal(err)
	}
	archiveHash, archiveSize, err := hashFileSHA256(run.Paths.OriginalArchive)
	if err != nil {
		t.Fatal(err)
	}
	extracted, err := safeExtractSourceZip(run.Paths.OriginalArchive, run.Paths.ImmutableSource)
	if err != nil {
		t.Fatal(err)
	}
	treeHash, files, extractedBytes, err := hashDirectoryTree(run.Paths.ImmutableSource)
	if err != nil {
		t.Fatal(err)
	}
	if files != extracted.FileCount || extractedBytes != extracted.ExtractedBytes {
		t.Fatalf("source identity mismatch: files=%d/%d bytes=%d/%d", files, extracted.FileCount, extractedBytes, extracted.ExtractedBytes)
	}
	run.Source = RepairSourceSnapshot{Filename: "source.zip", Size: archiveSize, SHA256: archiveHash, TreeSHA256: treeHash, FileCount: files, ExtractedBytes: extractedBytes, ImportedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := resetWorkingCopy(run); err != nil {
		t.Fatal(err)
	}
	profile, err := detectRepairProject(run.Paths.ImmutableSource)
	if err != nil {
		t.Fatal(err)
	}
	run.Project = profile
	run.State = "imported"
	run.Phase = "source-profiled"
	if err := writeRepairReceipt(run); err != nil {
		t.Fatal(err)
	}
	return app, run
}
