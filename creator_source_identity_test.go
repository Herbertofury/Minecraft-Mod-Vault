package main

import (
	"context"
	"testing"
)

func TestCreatorDescriptionRejectsSpoofedSources(t *testing.T) {
	for _, raw := range []string{
		"https://modrinth.com.attacker.invalid/mod/sodium",
		"https://notcurseforge.com/minecraft/mc-mods/create",
		"https://github.com.attacker.invalid/FabricMC/fabric",
		"https://modrinth.com@attacker.invalid/mod/sodium",
		"https://user:secret@modrinth.com/mod/sodium",
		"ftp://modrinth.com/mod/sodium",
		"https://modrinth.com:8443/mod/sodium",
		"https://modrinth.com/user/example",
		"https://www.curseforge.com/members/example/projects",
	} {
		if mod, ok := creatorModFromURL(raw, 0, CreatorVideo{}); ok {
			t.Errorf("accepted %q as %#v", raw, mod)
		}
	}
}

func TestCreatorDescriptionUsesProjectNotVersionIdentity(t *testing.T) {
	for _, tc := range []struct{ raw, id string }{
		{"https://modrinth.com/mod/sodium/version/abc?loader=fabric", "sodium"},
		{"https://modrinth.com/modpack/example/versions", "example"},
		{"https://www.curseforge.com/minecraft/mc-mods/create/files/123", "create"},
		{"https://www.curseforge.com/minecraft/texture-packs/example/files/all", "example"},
		{"https://github.com/Author/Project/releases/tag/v1", "Author/Project"},
	} {
		mod, ok := creatorModFromURL(tc.raw, 42, CreatorVideo{Platform: "youtube", URL: "https://www.youtube.com/watch?v=example"})
		if !ok || mod.ProjectID != tc.id || mod.TimestampS != 42 || mod.URL != tc.raw {
			t.Errorf("%q: %#v (ok=%v)", tc.raw, mod, ok)
		}
	}
}

func TestCreatorIdentityDoesNotMergeSameNamedProjects(t *testing.T) {
	a := CreatorMod{Name: "Example", Provider: "github", ProjectID: "Alice/Example"}
	b := CreatorMod{Name: "Example", Provider: "github", ProjectID: "Bob/Example"}
	if creatorResolvedIdentity(a) == creatorResolvedIdentity(b) {
		t.Fatal("distinct projects merged by title")
	}
	b = a
	b.Name = "Renamed Example"
	b.TimestampS = 50
	if creatorResolvedIdentity(a) != creatorResolvedIdentity(b) {
		t.Fatal("same project duplicated by name/timestamp")
	}
	b.ProjectID = "alice/Example"
	if creatorResolvedIdentity(a) == creatorResolvedIdentity(b) {
		t.Fatal("case-sensitive IDs collapsed")
	}
}

func TestCreatorAnalyzeRejectsLookalikeBeforeNetwork(t *testing.T) {
	app := &App{} // no HTTP client: rejected inputs must never reach network.
	for _, raw := range []string{"https://youtube.com.attacker.invalid/watch?v=a", "https://nottiktok.com/video/1", "file://youtube.com/watch?v=a", "https://user:secret@youtube.com/watch?v=a"} {
		if _, err := app.analyzeCreatorVideo(context.Background(), raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}
