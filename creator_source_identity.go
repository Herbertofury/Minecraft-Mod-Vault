package main

import (
	"net/url"
	"strings"
)

func creatorPublicWebURL(u *url.URL) bool {
	return u != nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "" && u.User == nil && (u.Port() == "" || u.Port() == "443" || u.Port() == "80")
}

func creatorModrinthKind(kind string) bool {
	switch kind {
	case "mod", "modpack", "plugin", "resourcepack", "shader", "datapack":
		return true
	}
	return false
}

func creatorCurseForgeKind(kind string) bool {
	switch kind {
	case "mc-mods", "modpacks", "texture-packs", "shaders", "customization", "worlds", "bukkit-plugins":
		return true
	}
	return false
}

// Titles are display text, not identity. Distinct projects may have identical
// names; one project may be linked at several versions or timestamps.
func creatorResolvedIdentity(mod CreatorMod) string {
	provider := strings.ToLower(strings.TrimSpace(mod.Provider))
	if id := strings.TrimSpace(mod.ProjectID); id != "" {
		// Provider IDs may be case-sensitive. Preserve them rather than combining
		// two different projects on a speculative normalization.
		return provider + ":id:" + id
	}
	if raw := strings.TrimSpace(mod.URL); raw != "" {
		return provider + ":url:" + raw
	}
	return provider + ":name:" + normalizeProjectTitle(mod.Name)
}
