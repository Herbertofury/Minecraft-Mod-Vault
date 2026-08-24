package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (a *App) installHangar(ctx context.Context, id, game, target string) (map[string]any, error) {
	base := providerBase("MMV_HANGAR_API_BASE", "https://hangar.papermc.io/api/v1")
	slug := id
	if strings.Contains(slug, "/") {
		slug = strings.SplitN(slug, "/", 2)[1]
	}
	a.mu.RLock()
	platform := strings.ToUpper(strings.TrimSpace(a.settings.ServerPlatform))
	a.mu.RUnlock()
	if platform == "" {
		platform = "PAPER"
	}
	var vr struct {
		Result []map[string]any `json:"result"`
	}
	if err := a.getJSON(ctx, base+"/projects/"+url.PathEscape(slug)+"/versions?limit=25&offset=0", nil, &vr); err != nil {
		return nil, err
	}
	if len(vr.Result) == 0 {
		return nil, errors.New("Hangar did not return any published versions")
	}
	type candidate struct {
		versionID string
		version   string
		url       string
		name      string
		size      int64
		sha256    string
		gameOK    bool
	}
	cands := []candidate{}
	for _, vm := range vr.Result {
		downloads, _ := vm["downloads"].(map[string]any)
		draw, ok := lookupFold(downloads, platform)
		if !ok {
			continue
		}
		dm, _ := draw.(map[string]any)
		if dm == nil {
			continue
		}
		dl := firstNonEmpty(stringFromAny(dm["downloadUrl"]), stringFromAny(dm["externalUrl"]))
		if dl == "" {
			continue
		}
		fi, _ := dm["fileInfo"].(map[string]any)
		name := ""
		size := int64(0)
		sha256 := ""
		if fi != nil {
			name = stringFromAny(fi["name"])
			size = int64FromAny(fi["sizeBytes"])
			sha256 = stringFromAny(fi["sha256Hash"])
		}
		if name == "" {
			name = safeFilename(slug + "-" + firstNonEmpty(stringFromAny(vm["name"]), stringFromAny(vm["id"])) + ".jar")
		}
		gameOK := game == ""
		if deps, ok := vm["platformDependencies"].(map[string]any); ok {
			if raw, ok := lookupFold(deps, platform); ok {
				versions := stringSliceFromAny(raw)
				if len(versions) == 0 || containsFold(versions, game) {
					gameOK = true
				}
			}
		} else {
			gameOK = true
		}
		cands = append(cands, candidate{versionID: stringFromAny(vm["id"]), version: stringFromAny(vm["name"]), url: absoluteURL("https://hangar.papermc.io", dl), name: name, size: size, sha256: sha256, gameOK: gameOK})
	}
	if len(cands) == 0 {
		return nil, fmt.Errorf("no Hangar %s download exists for %s", platform, slug)
	}
	chosen := cands[0]
	if game != "" {
		found := false
		for _, c := range cands {
			if c.gameOK {
				chosen, found = c, true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Hangar has no %s build explicitly compatible with Minecraft %s", platform, game)
		}
	}
	if target == "" || target == "auto" {
		target = "plugins"
	}
	if target != "plugins" && target != "downloads" {
		return nil, errors.New("Hangar projects are server plugins; choose the plugins or downloads target")
	}
	dir := a.javaTargetDir(target)
	if err := osMkdirAll(dir); err != nil {
		return nil, err
	}
	dst := uniquePath(filepath.Join(dir, safeFilename(chosen.name)))
	hashes := map[string]string{}
	if chosen.sha256 != "" {
		hashes["sha256"] = chosen.sha256
	}
	if err := a.downloadURLVerified(ctx, chosen.url, dst, chosen.size, hashes); err != nil {
		return nil, err
	}
	if strings.HasSuffix(strings.ToLower(dst), ".jar") {
		if err := validateZipContainer(dst); err != nil {
			_ = os.Remove(dst)
			return nil, fmt.Errorf("downloaded Hangar plugin is not a valid JAR: %w", err)
		}
	}
	return map[string]any{"ok": true, "provider": "hangar", "project": slug, "version": chosen.version, "versionId": chosen.versionID, "platform": platform, "file": filepath.Base(dst), "path": dst, "target": target}, nil
}

func (a *App) installSpigot(ctx context.Context, id, target string) (map[string]any, error) {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return nil, errors.New("Spigot resource ID must be numeric")
	}
	base := providerBase("MMV_SPIGET_API_BASE", "https://api.spiget.org/v2")
	var resource map[string]any
	if err := a.getJSON(ctx, base+"/resources/"+url.PathEscape(id), nil, &resource); err != nil {
		return nil, err
	}
	if boolFromAny(resource["external"]) {
		return nil, errors.New("this Spigot resource uses an external download; the Vault keeps its details integrated but will not guess an external package")
	}
	if target == "" || target == "auto" {
		target = "plugins"
	}
	if target != "plugins" && target != "downloads" {
		return nil, errors.New("Spigot resources install to plugins or Vault downloads")
	}
	name := firstNonEmpty(stringFromAny(resource["name"]), "spigot-resource-"+id)
	filename := safeFilename(regexpSafeFilename(name) + ".jar")
	dir := a.javaTargetDir(target)
	if err := osMkdirAll(dir); err != nil {
		return nil, err
	}
	dst := uniquePath(filepath.Join(dir, filename))
	dl := base + "/resources/" + url.PathEscape(id) + "/download"
	if err := a.downloadURLVerified(ctx, dl, dst, 0, nil); err != nil {
		return nil, err
	}
	if err := validateZipContainer(dst); err != nil {
		_ = os.Remove(dst)
		return nil, fmt.Errorf("Spiget download was not a valid JAR: %w", err)
	}
	return map[string]any{"ok": true, "provider": "spigot", "project": id, "file": filepath.Base(dst), "path": dst, "target": target}, nil
}

func (a *App) installBuiltByBit(ctx context.Context, id, target string) (map[string]any, error) {
	a.mu.RLock()
	token := strings.TrimSpace(a.settings.BuiltByBitOAuthToken)
	a.mu.RUnlock()
	if token == "" {
		return nil, errors.New("BuiltByBit one-click licensed downloads need an OAuth access token in Settings; search and full project details use the API token separately")
	}
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return nil, errors.New("BuiltByBit resource ID must be numeric")
	}
	base := providerBase("MMV_BUILTBYBIT_API_BASE", "https://api.builtbybit.com")
	headers := map[string]string{"Authorization": "Bearer " + token}
	var initResp struct {
		Data struct {
			Request struct {
				Token string `json:"token"`
			} `json:"request"`
		} `json:"data"`
	}
	initURL := base + "/v2/resources/discover/download/direct/initiate?" + url.Values{"content_type": {"resource"}, "content_id": {id}}.Encode()
	if err := a.getJSON(ctx, initURL, headers, &initResp); err != nil {
		return nil, err
	}
	if initResp.Data.Request.Token == "" {
		return nil, errors.New("BuiltByBit did not return a download request token")
	}
	var downloadURL string
	for attempt := 0; attempt < 12; attempt++ {
		var statusResp struct {
			Data struct {
				Status struct {
					Retry bool   `json:"retry"`
					URL   string `json:"url"`
				} `json:"status"`
			} `json:"data"`
		}
		statusURL := base + "/v2/resources/discover/download/direct/status?" + url.Values{"token": {initResp.Data.Request.Token}}.Encode()
		if err := a.getJSON(ctx, statusURL, headers, &statusResp); err != nil {
			return nil, err
		}
		if statusResp.Data.Status.URL != "" {
			downloadURL = statusResp.Data.Status.URL
			break
		}
		if !statusResp.Data.Status.Retry {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(750 * time.Millisecond):
		}
	}
	if downloadURL == "" {
		return nil, errors.New("BuiltByBit download request did not produce a downloadable file")
	}
	if target == "" || target == "auto" {
		target = "downloads"
	}
	dir := a.javaTargetDir(target)
	if err := osMkdirAll(dir); err != nil {
		return nil, err
	}
	filename := "builtbybit-resource-" + id + downloadExtension(downloadURL)
	dst := uniquePath(filepath.Join(dir, filename))
	if err := a.downloadURLVerified(ctx, downloadURL, dst, 0, nil); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "provider": "builtbybit", "project": id, "file": filepath.Base(dst), "path": dst, "target": target}, nil
}

func lookupFold(m map[string]any, key string) (any, bool) {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

func validateZipContainer(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	if len(r.File) == 0 {
		return errors.New("archive has no entries")
	}
	return nil
}

func regexpSafeFilename(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.', r == ' ':
			return r
		default:
			return '-'
		}
	}, s)
	s = strings.Trim(strings.Join(strings.Fields(s), "-"), "-.")
	if s == "" {
		return "plugin"
	}
	return s
}

func downloadExtension(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		if ext := strings.ToLower(filepath.Ext(u.Path)); ext != "" && len(ext) <= 10 {
			return ext
		}
	}
	return ".bin"
}
