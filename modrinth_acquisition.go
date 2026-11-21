package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// A graph in one native acquisition request, not a profile-wide resolver.
// All provider metadata is checked before the first package or directory write.
type modrinthAcquisition struct {
	project    ModrinthProject
	version    ModrinthVersion
	file       ModrinthVersionFile
	target     string
	dependency bool
}

func (a *App) planModrinthAcquisition(ctx context.Context, projectID, exactVersionID, game, loader, target string, dependency bool) ([]modrinthAcquisition, error) {
	const maxProjects, maxEdges = 64, 512
	projects := map[string]ModrinthProject{}
	exact := map[string]ModrinthVersion{}
	selected := map[string]modrinthAcquisition{}
	active := map[string]bool{}
	explicit := map[string]bool{}
	order := []modrinthAcquisition{}
	edges := 0
	var resolve func(string, string, string, bool) error
	resolve = func(id, pin, destination string, child bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var version ModrinthVersion
		if pin != "" {
			version = exact[pin]
			if version.ID == "" {
				var err error
				version, err = a.fetchModrinthVersion(ctx, pin)
				if err != nil {
					return err
				}
				if version.ID != pin || version.ProjectID == "" {
					return fmt.Errorf("exact version %q returned unconfirmed version or project identity", pin)
				}
				exact[pin] = version
			}
			if id == "" {
				id = version.ProjectID
			}
		}
		if id == "" {
			return fmt.Errorf("required dependency has no native project or exact version; review its publisher instructions")
		}
		project := projects[id]
		if project.ID == "" {
			if len(selected) >= maxProjects {
				return fmt.Errorf("requirements exceed %d projects; review the acquisition in smaller sets", maxProjects)
			}
			var err error
			project, err = a.fetchModrinthProject(ctx, id)
			if err != nil {
				return err
			}
			if project.ID == "" || (project.ID != id && project.Slug != id) {
				return fmt.Errorf("project %q returned unconfirmed canonical identity", id)
			}
			projects[id] = project
			projects[project.ID] = project
		}
		canonical := project.ID
		if pin != "" && version.ProjectID != canonical {
			return fmt.Errorf("exact version %q belongs to %q, not requested project %q", pin, version.ProjectID, canonical)
		}
		if prior, ok := selected[canonical]; ok {
			if pin != "" && prior.version.ID != pin {
				if explicit[canonical] {
					return fmt.Errorf("no common exact version for %s: requires both %s and %s", project.Title, prior.version.ID, pin)
				}
				return fmt.Errorf("%s was selected as %s but another dependency pins %s; review the exact selection", project.Title, prior.version.ID, pin)
			}
			if pin != "" {
				explicit[canonical] = true
			}
			if active[canonical] {
				return fmt.Errorf("required-dependency cycle at %s (%s)", project.Title, prior.version.ID)
			}
			return nil
		}
		if len(selected) >= maxProjects {
			return fmt.Errorf("requirements exceed %d projects; review the acquisition in smaller sets", maxProjects)
		}
		if pin == "" {
			versions, err := a.fetchModrinthVersions(ctx, canonical, game, loader, project.ProjectType)
			if err != nil {
				return err
			}
			if len(versions) == 0 {
				return fmt.Errorf("no compatible %s version was found for Minecraft %s / %s", project.Title, game, loader)
			}
			version = versions[0]
		}
		if version.ID == "" || version.ProjectID != canonical {
			return fmt.Errorf("%s returned a version without matching canonical project and version IDs", project.Title)
		}
		if (game != "" && !containsFold(version.GameVersions, game)) ||
			(project.ProjectType == "mod" && loader != "" && loader != "any" && loader != "vanilla" && !containsFold(version.Loaders, loader)) {
			return fmt.Errorf("%s version %s is not confirmed for Minecraft %s / %s", project.Title, version.ID, game, loader)
		}
		if len(version.Files) == 0 {
			return fmt.Errorf("version %s of %s has no downloadable files", version.ID, project.Title)
		}
		file := version.Files[0]
		for _, f := range version.Files {
			if f.Primary {
				file = f
				break
			}
		}
		u, err := url.Parse(file.URL)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" {
			return fmt.Errorf("version %s of %s has no HTTPS file route", version.ID, project.Title)
		}
		if destination == "" || destination == "auto" {
			switch project.ProjectType {
			case "mod":
				destination = "mods"
			case "resourcepack":
				destination = "resourcepacks"
			case "shader":
				destination = "shaderpacks"
			default:
				destination = "downloads"
			}
		}
		node := modrinthAcquisition{project: project, version: version, file: file, target: destination, dependency: child}
		selected[canonical] = node
		explicit[canonical] = pin != ""
		active[canonical] = true
		// Within one dependency list, prefer explicit pins before flexible project
		// picks. Across different branches this is deliberately not a search solver.
		for _, pinned := range []bool{true, false} {
			for _, dep := range version.Dependencies {
				if dep.DependencyType != "required" || (dep.VersionID != "") != pinned {
					continue
				}
				edges++
				if edges > maxEdges {
					return fmt.Errorf("requirements exceed %d edges; review the acquisition in smaller sets", maxEdges)
				}
				if dep.ProjectID == "" && dep.VersionID == "" {
					name := strings.TrimSpace(dep.FileName)
					if name == "" {
						name = "unnamed dependency"
					}
					return fmt.Errorf("%s requires %q without a native identity; review its publisher instructions", project.Title, name)
				}
				if err := resolve(dep.ProjectID, dep.VersionID, "auto", true); err != nil {
					return fmt.Errorf("required by %s: %w", project.Title, err)
				}
			}
		}
		delete(active, canonical)
		order = append(order, node)
		return nil
	}
	if err := resolve(projectID, exactVersionID, target, dependency); err != nil {
		return nil, err
	}
	return order, nil
}
