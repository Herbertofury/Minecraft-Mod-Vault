package main

import (
	"strings"
	"testing"
)

func TestVersionAtlasLoadsAuthoritativeCorpus(t *testing.T) {
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		t.Fatal(err)
	}
	if atlas.Summary.MojangVersions < 900 {
		t.Fatalf("Mojang corpus unexpectedly small: %d", atlas.Summary.MojangVersions)
	}
	if atlas.Summary.MCMetaVersions < 400 {
		t.Fatalf("mcmeta corpus unexpectedly small: %d", atlas.Summary.MCMetaVersions)
	}
	if atlas.Summary.RuntimeLibraries < 50000 {
		t.Fatalf("runtime-library corpus unexpectedly small: %d", atlas.Summary.RuntimeLibraries)
	}
	if atlas.Summary.LatestRelease == "" || atlas.Summary.LatestSnapshot == "" {
		t.Fatalf("latest identities missing: %#v", atlas.Summary)
	}
	for _, id := range []string{"fabric-loom", "forgegradle", "moddevgradle", "neoforge", "yarn", "intermediary"} {
		if _, ok := atlas.Maven[id]; !ok {
			t.Fatalf("required toolchain %q missing from atlas", id)
		}
	}
	if version, ok := atlas.ByID["1.20.1"]; !ok || version.JavaMajor == 0 || !version.HasClient || !version.HasServer {
		t.Fatalf("1.20.1 authoritative profile incomplete: %#v, present=%v", version, ok)
	}
}

func TestVersionAtlasRoutesLegacyAndModernForgeDifferently(t *testing.T) {
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		t.Fatal(err)
	}
	legacy := atlas.Resolve("1.12.2", "forge")
	if !legacy.Exists || !legacy.LoaderSupported {
		t.Fatalf("legacy Forge resolution incomplete: %#v", legacy)
	}
	if !hasAtlasChoice(legacy.CompatibilityRoutes, "retrofuturagradle") || !hasAtlasChoice(legacy.CompatibilityRoutes, "retromcp-java") {
		t.Fatalf("legacy recovery route missing: %#v", legacy.CompatibilityRoutes)
	}
	if hasAtlasChoice(legacy.BuildToolchains, "forgegradle") {
		t.Fatalf("legacy 1.12.2 must not be rewritten with a modern ForgeGradle lane: %#v", legacy.BuildToolchains)
	}

	modern := atlas.Resolve("1.20.1", "forge")
	if !modern.Exists || !modern.LoaderSupported {
		t.Fatalf("modern Forge resolution incomplete: %#v", modern)
	}
	if !hasAtlasChoice(modern.BuildToolchains, "forgegradle") {
		t.Fatalf("modern ForgeGradle lane missing: %#v", modern.BuildToolchains)
	}
	for _, choice := range modern.CompatibilityRoutes {
		if strings.Contains(strings.ToLower(choice.ID), "retro") {
			t.Fatalf("modern Forge resolution leaked legacy route: %#v", modern.CompatibilityRoutes)
		}
	}
}

func TestVersionAtlasResolvesFabricWithExactTargetEvidence(t *testing.T) {
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		t.Fatal(err)
	}
	resolution := atlas.Resolve("1.21.1", "fabric")
	if !resolution.Exists || !resolution.LoaderSupported {
		t.Fatalf("Fabric target not supported: %#v", resolution)
	}
	if resolution.JavaMajor < 21 {
		t.Fatalf("unexpected Java requirement: %d", resolution.JavaMajor)
	}
	if resolution.LoaderVersion == "" || !hasAtlasChoice(resolution.BuildToolchains, "fabric-loom") {
		t.Fatalf("Fabric loader/toolchain not resolved: %#v", resolution)
	}
	if !hasAtlasChoice(resolution.Mappings, "intermediary") {
		t.Fatalf("exact intermediary mapping missing: %#v", resolution.Mappings)
	}
}

func hasAtlasChoice(rows []AtlasToolchainChoice, id string) bool {
	for _, row := range rows {
		if row.ID == id && (row.Version != "" || row.Channel == "specialized") {
			return true
		}
	}
	return false
}
