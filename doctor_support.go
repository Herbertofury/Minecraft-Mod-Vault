package main

import (
	"regexp"
	"strings"
)

func doctorMinecraftConstraintLikelyMatches(constraint, target string) (bool, bool) {
	constraint = strings.TrimSpace(strings.Trim(constraint, "\""))
	target = strings.TrimSpace(target)
	if constraint == "" || constraint == "<nil>" || constraint == "*" || target == "" {
		return true, false
	}
	for _, alternative := range strings.Split(constraint, "||") {
		alternative = strings.TrimSpace(alternative)
		if alternative == "" {
			continue
		}
		if ok, known := doctorMinecraftConstraintConjunction(alternative, target); known && ok {
			return true, true
		}
	}
	if strings.Contains(constraint, "||") {
		return false, true
	}
	return doctorMinecraftConstraintConjunction(constraint, target)
}

func doctorMinecraftConstraintConjunction(constraint, target string) (bool, bool) {
	constraint = strings.TrimSpace(constraint)
	if match := regexp.MustCompile(`^([\[\(])\s*([^,]*)\s*,\s*([^\]\)]*)\s*([\]\)])$`).FindStringSubmatch(constraint); len(match) == 5 {
		lowerOK, upperOK := true, true
		if strings.TrimSpace(match[2]) != "" {
			cmp := compareGameVersions(target, strings.TrimSpace(match[2]))
			lowerOK = cmp > 0 || (cmp == 0 && match[1] == "[")
		}
		if strings.TrimSpace(match[3]) != "" {
			cmp := compareGameVersions(target, strings.TrimSpace(match[3]))
			upperOK = cmp < 0 || (cmp == 0 && match[4] == "]")
		}
		return lowerOK && upperOK, true
	}
	parts := strings.Fields(strings.NewReplacer(",", " ", "&&", " ").Replace(constraint))
	if len(parts) > 1 {
		known := false
		for _, part := range parts {
			ok, partKnown := doctorMinecraftConstraintConjunction(part, target)
			known = known || partKnown
			if partKnown && !ok {
				return false, true
			}
		}
		return true, known
	}
	for _, operator := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(constraint, operator) {
			version := strings.TrimSpace(strings.TrimPrefix(constraint, operator))
			if !parseGameVersionKey(version).Valid {
				return true, false
			}
			cmp := compareGameVersions(target, version)
			switch operator {
			case ">=":
				return cmp >= 0, true
			case "<=":
				return cmp <= 0, true
			case ">":
				return cmp > 0, true
			case "<":
				return cmp < 0, true
			default:
				return cmp == 0, true
			}
		}
	}
	if strings.ContainsAny(constraint, "*xX") {
		prefix := strings.TrimRight(constraint, ".*xX")
		return strings.HasPrefix(target, prefix), true
	}
	if strings.HasPrefix(constraint, "~") {
		base := strings.TrimPrefix(constraint, "~")
		baseKey, targetKey := parseGameVersionKey(base), parseGameVersionKey(target)
		return baseKey.Valid && targetKey.Valid && baseKey.Parts[0] == targetKey.Parts[0] && baseKey.Parts[1] == targetKey.Parts[1] && compareGameVersions(target, base) >= 0, true
	}
	if strings.HasPrefix(constraint, "^") {
		base := strings.TrimPrefix(constraint, "^")
		baseKey, targetKey := parseGameVersionKey(base), parseGameVersionKey(target)
		return baseKey.Valid && targetKey.Valid && baseKey.Parts[0] == targetKey.Parts[0] && compareGameVersions(target, base) >= 0, true
	}
	if parseGameVersionKey(constraint).Valid {
		return compareGameVersions(constraint, target) == 0, true
	}
	return true, false
}
