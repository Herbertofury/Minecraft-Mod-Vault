package main

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

// ProviderInfo is the server-owned description of a source lane. The UI consumes
// this instead of carrying its own hard-coded provider list, so provider coverage
// can grow without leaving stale controls behind.
type ProviderInfo struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Group                string   `json:"group"`
	Description          string   `json:"description"`
	HomeURL              string   `json:"homeUrl,omitempty"`
	SearchMode           string   `json:"searchMode"`
	DetailMode           string   `json:"detailMode"`
	InstallMode          string   `json:"installMode"`
	Credential           string   `json:"credential,omitempty"`
	CredentialConfigured bool     `json:"credentialConfigured"`
	ProjectTypes         []string `json:"projectTypes,omitempty"`
	Capabilities         []string `json:"capabilities,omitempty"`
	DefaultEnabled       bool     `json:"defaultEnabled"`
}

const providerSchemaVersion = 6

type ProviderHealth struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	LastAttempt   string `json:"lastAttempt,omitempty"`
	LastSuccess   string `json:"lastSuccess,omitempty"`
	LastError     string `json:"lastError,omitempty"`
	LastResultCnt int    `json:"lastResultCount,omitempty"`
	LatencyMS     int64  `json:"latencyMs,omitempty"`
}

var providerCatalog = []ProviderInfo{
	{ID: "modrinth", Name: "Modrinth", Group: "Major hubs", Description: "Native project search, rich project metadata, versions, dependencies, media and verified file installs.", HomeURL: "https://modrinth.com/", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "verified-native", ProjectTypes: []string{"mod", "modpack", "resourcepack", "shader", "plugin", "datapack"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "dependencies", "install", "update"}, DefaultEnabled: true},
	{ID: "curseforge", Name: "CurseForge", Group: "Major hubs", Description: "Native API when configured, with an integrated public-index fallback for browsing and project media.", HomeURL: "https://www.curseforge.com/minecraft", SearchMode: "api+web", DetailMode: "api+web", InstallMode: "verified-native-with-key", ProjectTypes: []string{"mod", "modpack", "resourcepack", "shader", "world", "addon", "skin", "tool", "datapack", "plugin"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "install", "update"}, DefaultEnabled: true},
	{ID: "github", Name: "GitHub", Group: "Source & releases", Description: "Repository and release discovery with owner avatars, release assets and verified direct release installation.", HomeURL: "https://github.com/", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "release-assets", ProjectTypes: []string{"mod", "plugin", "datapack", "resourcepack", "shader", "tool"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "install", "source"}, DefaultEnabled: true},
	{ID: "smithed", Name: "Smithed", Group: "Data & resource packs", Description: "Native Smithed v2 pack search with pack metadata, categories, galleries, Minecraft compatibility, versions, dependencies and direct datapack/resource-pack installation.", HomeURL: "https://smithed.dev/", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "native-pack-download", ProjectTypes: []string{"datapack", "resourcepack"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "dependencies", "install", "update", "datapacks"}, DefaultEnabled: true},
	{ID: "planetminecraft", Name: "Planet Minecraft", Group: "Community", Description: "Integrated community index for mods, data packs, texture packs, maps and CIT content, with in-app page enrichment.", HomeURL: "https://www.planetminecraft.com/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "detected-downloads", ProjectTypes: []string{"mod", "datapack", "resourcepack", "world", "skin"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection"}, DefaultEnabled: true},
	{ID: "mcpedl", Name: "MCPEDL", Group: "Bedrock", Description: "Integrated Bedrock add-on, resource-pack, map and furniture discovery with page media and package-link detection.", HomeURL: "https://mcpedl.com/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "detected-bedrock-package", ProjectTypes: []string{"addon", "resourcepack", "world", "shader"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "bedrock"}, DefaultEnabled: true},
	{ID: "marketplace", Name: "Minecraft Marketplace", Group: "Bedrock", Description: "Official Marketplace search normalized into the Vault with artwork, creator and product details kept inside the app.", HomeURL: "https://www.minecraft.net/en-us/marketplace", SearchMode: "official-web-index", DetailMode: "official-web-page", InstallMode: "minecraft-client-handoff", ProjectTypes: []string{"addon", "world", "resourcepack", "skin"}, Capabilities: []string{"search", "details", "media", "authors", "bedrock"}, DefaultEnabled: true},
	{ID: "hangar", Name: "Hangar", Group: "Plugins", Description: "PaperMC Hangar project search, project metadata, supported platforms, versions and version downloads through the Hangar API.", HomeURL: "https://hangar.papermc.io/", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "native-version-download", ProjectTypes: []string{"plugin"}, Capabilities: []string{"search", "details", "authors", "versions", "install", "update", "plugins"}, DefaultEnabled: true},
	{ID: "spigot", Name: "SpigotMC", Group: "Plugins", Description: "SpigotMC resource discovery normalized through the Spiget index, with in-app resource details and downloadable resources when available.", HomeURL: "https://www.spigotmc.org/resources/", SearchMode: "spiget-index", DetailMode: "spiget-index", InstallMode: "resource-download-when-available", ProjectTypes: []string{"plugin"}, Capabilities: []string{"search", "details", "authors", "versions", "download-detection", "plugins"}, DefaultEnabled: true},
	{ID: "bukkitdev", Name: "BukkitDev", Group: "Plugins", Description: "Integrated BukkitDev project index, kept separate from CurseForge so Bukkit-first projects remain searchable as their own lane.", HomeURL: "https://dev.bukkit.org/bukkit-plugins", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "detected-downloads", ProjectTypes: []string{"plugin"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "plugins"}, DefaultEnabled: true},
	{ID: "spongeore", Name: "Sponge Ore", Group: "Plugins", Description: "Sponge Ore plugin catalog integrated through the Ore project API and project pages for Sponge server plugins.", HomeURL: "https://ore.spongepowered.org/", SearchMode: "native-api", DetailMode: "native-api+web", InstallMode: "native-version-download", ProjectTypes: []string{"plugin"}, Capabilities: []string{"search", "details", "authors", "versions", "install", "plugins"}, DefaultEnabled: true},
	{ID: "builtbybit", Name: "BuiltByBit", Group: "Marketplaces", Description: "Official Discovery API integration for resources, creators, images, categories, reviews and versions when an API token is configured.", HomeURL: "https://builtbybit.com/resources/", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "licensed-download-api", Credential: "builtbybit", ProjectTypes: []string{"plugin", "mod", "resourcepack", "tool", "server"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "reviews", "download-api"}, DefaultEnabled: false},
	{ID: "polymart", Name: "Polymart", Group: "Marketplaces", Description: "Integrated Polymart product browser for plugins, mods, resource packs, models, builds and server assets, with project pages rendered inside the Vault.", HomeURL: "https://polymart.org/products", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "integrated-metadata", ProjectTypes: []string{"plugin", "mod", "resourcepack", "tool", "server", "world"}, Capabilities: []string{"search", "details", "media", "authors", "marketplace"}, DefaultEnabled: true},
	{ID: "minecraftmaps", Name: "Minecraft Maps", Group: "Maps & worlds", Description: "MinecraftMaps.com map discovery normalized into the Vault with map art, creators, compatibility clues and verified detected-package installation when a real archive is exposed.", HomeURL: "https://www.minecraftmaps.com/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"world"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "worlds"}, DefaultEnabled: true},
	{ID: "resourcepacknet", Name: "ResourcePack.net", Group: "Data & resource packs", Description: "ResourcePack.net texture/resource-pack index integrated into Vault search and rich project details, with verified detected archive installs when available.", HomeURL: "https://resourcepack.net/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"resourcepack"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "resourcepacks"}, DefaultEnabled: true},
	{ID: "texturepacks", Name: "Texture-Packs.com", Group: "Data & resource packs", Description: "Texture-Packs.com resource-pack discovery integrated into the Vault with imagery, compatibility clues, in-app details and verified detected archive installs when available.", HomeURL: "https://texture-packs.com/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"resourcepack"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "resourcepacks"}, DefaultEnabled: true},
	{ID: "moddb", Name: "Mod DB", Group: "Community", Description: "Minecraft Mod DB projects integrated through its public project index and project pages.", HomeURL: "https://www.moddb.com/games/minecraft/mods", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "detected-downloads", ProjectTypes: []string{"mod", "modpack", "tool"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection"}, DefaultEnabled: true},
	{ID: "atlauncher", Name: "ATLauncher", Group: "Modpacks", Description: "Public ATLauncher GraphQL modpack search with current Minecraft version and pack-version details.", HomeURL: "https://atlauncher.com/packs/all", SearchMode: "native-graphql", DetailMode: "native-graphql", InstallMode: "modpack-manifest", ProjectTypes: []string{"modpack"}, Capabilities: []string{"search", "details", "versions", "modpacks"}, DefaultEnabled: true},
	{ID: "technic", Name: "Technic", Group: "Modpacks", Description: "Technic Platform modpack search and pack metadata through the launcher's public Platform API.", HomeURL: "https://www.technicpack.net/modpacks", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "modpack-manifest", ProjectTypes: []string{"modpack"}, Capabilities: []string{"search", "details", "media", "versions", "modpacks"}, DefaultEnabled: true},
	{ID: "ftb", Name: "Feed The Beast", Group: "Modpacks", Description: "FTB/modpacks.ch public catalog integration for current packs, versions and Minecraft compatibility.", HomeURL: "https://www.feed-the-beast.com/modpacks", SearchMode: "native-public-api", DetailMode: "native-public-api", InstallMode: "modpack-manifest", ProjectTypes: []string{"modpack"}, Capabilities: []string{"search", "details", "media", "versions", "modpacks"}, DefaultEnabled: true},
	{ID: "nexusmods", Name: "Nexus Mods", Group: "Community", Description: "Nexus Mods Minecraft metadata through the Nexus API when an API key is configured.", HomeURL: "https://www.nexusmods.com/minecraft", SearchMode: "native-api", DetailMode: "native-api", InstallMode: "integrated-metadata", Credential: "nexus", ProjectTypes: []string{"mod", "resourcepack", "tool"}, Capabilities: []string{"search", "details", "media", "authors", "versions"}, DefaultEnabled: false},
	{ID: "vanillatweaks", Name: "Vanilla Tweaks", Group: "Data & resource packs", Description: "Official Vanilla Tweaks resource-pack, data-pack and crafting-tweak pickers searchable inside the Vault, with the selected official picker kept in-app.", HomeURL: "https://vanillatweaks.net/", SearchMode: "official-web-index", DetailMode: "official-web-page", InstallMode: "official-picker", ProjectTypes: []string{"resourcepack", "datapack"}, Capabilities: []string{"search", "details", "media", "datapacks"}, DefaultEnabled: true},
	{ID: "minecrafthub", Name: "MinecraftHub", Group: "Curated discovery", Description: "Curated cross-provider Minecraft directory integrated as a native Vault lane. Search, compatibility, creator and media metadata stay in-app, while installs resolve back to the original provider and use Vault’s verified provider installer instead of treating the directory as a mirror.", HomeURL: "https://minecrafthub.io/resources", SearchMode: "curated-cross-provider-index", DetailMode: "integrated-rich-page", InstallMode: "canonical-provider-resolution", ProjectTypes: []string{"mod", "resourcepack", "shader", "addon", "modpack", "datapack", "plugin", "world", "skin"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "compatibility", "source-resolution", "install"}, DefaultEnabled: true},
	{ID: "mcreator", Name: "MCreator Community", Group: "Community", Description: "Live MCreator modification catalog for community-made Java mods and Bedrock add-ons, with in-app project pages and verified detected package installation.", HomeURL: "https://mcreator.net/modifications", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"mod", "addon"}, Capabilities: []string{"search", "details", "media", "authors", "versions", "download-detection"}, DefaultEnabled: true},
	{ID: "shaderpackscom", Name: "ShaderPacks.com", Group: "Shaders & visuals", Description: "Dedicated modern shader catalog normalized into Vault with screenshots, version clues and verified archive detection when the project exposes a package.", HomeURL: "https://shaderpacks.com/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"shader"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "shaders"}, DefaultEnabled: true},
	{ID: "shaderpacksnet", Name: "Shaderpacks.net", Group: "Shaders & visuals", Description: "Long-running shader directory integrated as a searchable Vault lane with rich project pages and verified final-archive detection.", HomeURL: "https://shaderpacks.net/category/shaderpacks/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"shader"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "shaders"}, DefaultEnabled: true},
	{ID: "minecraftshader", Name: "MinecraftShader.com", Group: "Shaders & visuals", Description: "Shader and visual-pack directory searched inside Vault, with project imagery, compatibility clues and validated package detection when available.", HomeURL: "https://minecraftshader.com/category/minecraft-shaders/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "verified-detected-download", ProjectTypes: []string{"shader", "resourcepack"}, Capabilities: []string{"search", "details", "media", "authors", "download-detection", "shaders"}, DefaultEnabled: true},
	{ID: "skindex", Name: "The Skindex", Group: "Skins & cosmetics", Description: "MinecraftSkins.com skin search rendered natively inside Vault with creator/project imagery and original PNG saving with Minecraft skin-dimension verification.", HomeURL: "https://www.minecraftskins.com/", SearchMode: "integrated-web-index", DetailMode: "integrated-web-page", InstallMode: "validated-skin-png", ProjectTypes: []string{"skin"}, Capabilities: []string{"search", "details", "media", "authors", "skin-download"}, DefaultEnabled: true},
}

func allProviderIDs(defaultOnly bool) []string {
	out := make([]string, 0, len(providerCatalog))
	for _, p := range providerCatalog {
		if defaultOnly && !p.DefaultEnabled {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

func providerInfoByID(id string) *ProviderInfo {
	id = strings.ToLower(strings.TrimSpace(id))
	for i := range providerCatalog {
		if providerCatalog[i].ID == id {
			return &providerCatalog[i]
		}
	}
	return nil
}

func (a *App) providerCredentialConfigured(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	switch id {
	case "curseforge":
		return strings.TrimSpace(a.settings.CurseForgeAPIKey) != ""
	case "github":
		return strings.TrimSpace(a.settings.GitHubToken) != ""
	case "builtbybit":
		return strings.TrimSpace(a.settings.BuiltByBitAPIKey) != ""
	case "nexusmods":
		return strings.TrimSpace(a.settings.NexusAPIKey) != ""
	default:
		return true
	}
}

func (a *App) providerInfos() []ProviderInfo {
	out := make([]ProviderInfo, len(providerCatalog))
	copy(out, providerCatalog)
	for i := range out {
		out[i].CredentialConfigured = a.providerCredentialConfigured(out[i].ID)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Group == out[j].Group {
			return out[i].Name < out[j].Name
		}
		return out[i].Group < out[j].Group
	})
	return out
}

func (a *App) enabledProvidersForType(projectType string) []string {
	a.mu.RLock()
	enabled := append([]string(nil), a.settings.EnabledSources...)
	a.mu.RUnlock()
	enabledSet := map[string]bool{}
	for _, id := range enabled {
		enabledSet[strings.ToLower(strings.TrimSpace(id))] = true
	}
	out := []string{}
	for _, p := range providerCatalog {
		if !enabledSet[p.ID] {
			continue
		}
		matched := false
		for _, t := range p.ProjectTypes {
			if projectTypeMatches(t, projectType) || strings.EqualFold(t, projectType) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		// Sources whose search itself requires credentials should not create
		// predictable background errors until they are configured.
		if p.Credential != "" && !a.providerCredentialConfigured(p.ID) {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

func (a *App) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.providerHealthMu.RLock()
	health := make(map[string]ProviderHealth, len(a.providerHealth))
	for k, v := range a.providerHealth {
		health[k] = v
	}
	a.providerHealthMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"providers": a.providerInfos(),
		"health":    health,
		"count":     len(providerCatalog),
		"refreshed": time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *App) noteProviderAttempt(id string, started time.Time, count int, err error) {
	now := time.Now().UTC()
	a.providerHealthMu.Lock()
	defer a.providerHealthMu.Unlock()
	if a.providerHealth == nil {
		a.providerHealth = map[string]ProviderHealth{}
	}
	h := a.providerHealth[id]
	h.ID = id
	h.LastAttempt = now.Format(time.RFC3339)
	h.LatencyMS = time.Since(started).Milliseconds()
	if err != nil {
		h.Status = "error"
		h.LastError = err.Error()
	} else {
		h.Status = "ok"
		h.LastError = ""
		h.LastSuccess = now.Format(time.RFC3339)
		h.LastResultCnt = count
	}
	a.providerHealth[id] = h
}
