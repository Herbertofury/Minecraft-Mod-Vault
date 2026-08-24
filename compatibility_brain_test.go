package main

import "testing"

func TestCompatibilityBrainImportsAndAnswersVersionQueries(t *testing.T) {
	brain, err := openCompatibilityBrain(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer brain.Close()

	status, err := brain.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready || status.MinecraftVersions < 900 || status.KnowledgeDocuments < 100 || status.ToolchainReleases < 1000 {
		t.Fatalf("compatibility brain corpus incomplete: %#v", status)
	}
	version, err := brain.MinecraftVersion("1.20.1")
	if err != nil {
		t.Fatal(err)
	}
	if !version.Available || !version.OfficialManifestKnown || version.JavaMajor == 0 || !version.HasClient || !version.HasServer {
		t.Fatalf("version intelligence incomplete: %#v", version)
	}
	missing, err := brain.MinecraftVersion("not-a-real-version")
	if err != nil {
		t.Fatal(err)
	}
	if missing.Available {
		t.Fatalf("unknown version unexpectedly available: %#v", missing)
	}
}

func TestCompatibilityBrainFullTextSearchFindsPortingKnowledge(t *testing.T) {
	brain, err := openCompatibilityBrain(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer brain.Close()

	rows, err := brain.Search("RetroFuturaGradle legacy Forge", "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("expected legacy Forge knowledge search results")
	}
	found := false
	for _, row := range rows {
		if row.Name == "RetroFuturaGradle" || row.ID == "retrofuturagradle" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("RetroFuturaGradle missing from search results: %#v", rows)
	}
}
