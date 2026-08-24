package main

import (
	"os"
	"strings"
	"testing"
)

func mustReadOmniAsset(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func TestOmniManagerPremiumAssetsAreWired(t *testing.T) {
	index := mustReadOmniAsset(t, "web/index.html")
	for _, want := range []string{
		`data-view="manager"`,
		`omnimanager.css`,
		`omnimanager-polish.js`,
		`Bedrock`,
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("manager shell missing %q", want)
		}
	}

	app := mustReadOmniAsset(t, "web/app.js")
	if !strings.Contains(app, "OmniManager") || !strings.Contains(app, "OmniManagerPolish") {
		t.Fatal("manager navigation is not connected to the OmniManager runtime")
	}

	css := mustReadOmniAsset(t, "web/omnimanager.css")
	for _, want := range []string{
		`.omni-layout`,
		`.omni-library`,
		`.omni-card`,
		`.omni-row`,
		`.omni-provider-stack`,
		`.omni-trust`,
		`prefers-reduced-motion`,
		`forced-colors`,
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("premium manager stylesheet missing %q", want)
		}
	}
	for _, forbidden := range []string{"content-visibility", "contain-intrinsic-size"} {
		if strings.Contains(css, forbidden) {
			t.Fatalf("manager stylesheet must not hide or defer installed content with %q", forbidden)
		}
	}
}

func TestOmniManagerHasNoFeatureTheater(t *testing.T) {
	index := strings.ToLower(mustReadOmniAsset(t, "web/index.html"))
	controller := strings.ToLower(mustReadOmniAsset(t, "web/omnimanager-polish.js"))
	for _, forbidden := range []string{"coming soon", "not implemented", "todo"} {
		if strings.Contains(index, forbidden) || strings.Contains(controller, forbidden) {
			t.Fatalf("manager surface contains forbidden placeholder language %q", forbidden)
		}
	}
}
