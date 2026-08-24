package main

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

const doctorMaxLogBytes = 20 << 20

type doctorLogRule struct {
	ID              string
	Severity        string
	Category        string
	Title           string
	Action          string
	Needles         []string
	SourceIDs       []string
	RepairPatternID string
	KnownRepair     bool
}

func (a *App) handleDoctorLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request DoctorLogRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	if strings.TrimSpace(request.Text) == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "paste a crash report, latest.log, debug.log, or build failure"})
		return
	}
	if len(request.Text) > doctorMaxLogBytes {
		writeJSON(w, http.StatusRequestEntityTooLarge, APIError{Error: "log exceeds the 20 MiB analysis limit"})
		return
	}
	if request.GameVersion == "" || request.Loader == "" {
		a.mu.RLock()
		settings := a.settings
		a.mu.RUnlock()
		request.GameVersion = firstNonEmpty(strings.TrimSpace(request.GameVersion), settings.GameVersion)
		request.Loader = normalizeDoctorLoader(firstNonEmpty(strings.TrimSpace(request.Loader), settings.Loader))
	}
	report, err := doctorAnalyzeLog(request)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func doctorAnalyzeLog(request DoctorLogRequest) (DoctorLogReport, error) {
	payload, err := loadDoctorToolsPayload()
	if err != nil {
		return DoctorLogReport{}, err
	}
	text := strings.ReplaceAll(request.Text, "\r\n", "\n")
	lower := strings.ToLower(text)
	lines := strings.Split(text, "\n")
	findings := []DoctorFinding{}
	evidenceLines := []string{}
	patternIDs := []string{}
	knownRepair := false
	for _, rule := range doctorLogRules() {
		if !doctorContainsAny(lower, rule.Needles...) {
			continue
		}
		matched := doctorMatchingLogLines(lines, rule.Needles, 6)
		evidenceLines = append(evidenceLines, matched...)
		findings = append(findings, DoctorFinding{
			ID:        rule.ID,
			Code:      rule.ID,
			Severity:  rule.Severity,
			Category:  rule.Category,
			Title:     rule.Title,
			Evidence:  firstNonEmpty(strings.Join(matched, " | "), "Signature found in supplied log"),
			Action:    rule.Action,
			SourceIDs: uniqueNonEmpty(rule.SourceIDs),
		})
		if rule.RepairPatternID != "" {
			patternIDs = appendUniqueString(patternIDs, rule.RepairPatternID)
		}
		knownRepair = knownRepair || rule.KnownRepair
	}
	if len(findings) == 0 {
		findings = append(findings, DoctorFinding{
			ID:       "unclassified-log",
			Code:     "unclassified-log",
			Severity: "warning",
			Category: "analysis",
			Title:    "No high-confidence built-in failure signature matched",
			Evidence: doctorFirstNonEmptyLogLine(lines),
			Action:   "Use the complete log, exact mod list, loader, Minecraft and Java versions. Follow the earliest causal exception through the final fatal chain, then correlate it with the JAR scan and dependency graph.",
			SourceIDs: []string{
				"mclogs", "crash-assistant", "minecraft-wiki",
			},
		})
	}
	root := doctorSelectRootCause(findings)
	selectedPatterns := []DoctorRepairPattern{}
	for _, pattern := range payload.RepairPatterns {
		for _, id := range patternIDs {
			if pattern.ID == id {
				selectedPatterns = append(selectedPatterns, pattern)
				break
			}
		}
	}
	sortDoctorFindings(findings)
	confidence := "medium"
	if knownRepair {
		confidence = "high"
	} else if root.ID == "unclassified-log" {
		confidence = "low"
	}
	return DoctorLogReport{
		CreatedAt:            time.Now().UTC().Format(time.RFC3339),
		Filename:             request.Filename,
		GameVersion:          request.GameVersion,
		Loader:               normalizeDoctorLoader(request.Loader),
		RootCause:            root,
		Findings:             findings,
		EvidenceLines:        uniqueNonEmpty(evidenceLines),
		RepairPatterns:       selectedPatterns,
		RecommendedSourceIDs: collectDoctorSourceIDs(findings, nil, nil),
		Confidence:           confidence,
	}, nil
}

func doctorLogRules() []doctorLogRule {
	return []doctorLogRule{
		{
			ID: "known-cataclysm-renderer-api-drift", Severity: "critical", Category: "binary-api", KnownRepair: true,
			Title:     "Known Cataclysm renderer owner drift matches a verified prior repair",
			Action:    "Verify the exact Cataclysm and Cataclysm Spellbooks builds. Reuse the recorded two-class owner/descriptor patch only when the old and replacement renderer class shapes match exactly.",
			Needles:   []string{"ancient_remnant_rework_renderer", "phantomancientremnantrenderer"},
			SourceIDs: []string{"repair-brain", "recaf", "asm", "vineflower"}, RepairPatternID: "binary-owner-descriptor-drift",
		},
		{
			ID: "known-realm-rpg-pool-namespace", Severity: "critical", Category: "data", KnownRepair: true,
			Title:     "Known Realm RPG jigsaw pool namespace defect matches prior evidence",
			Action:    "Inspect the exact NBT pool fields and sibling identifiers. Rewrite only the proven minecraft:<Realm-RPG-pool> references to the existing mod namespace, then validate structure generation on a copied world.",
			Needles:   []string{"minecraft:<realm-rpg-pool>", "realm-rpg-pool", "unknown registry key in resourcekey[minecraft:worldgen/template_pool"},
			SourceIDs: []string{"repair-brain", "nbt-studio", "amulet", "mcaselector"}, RepairPatternID: "data-namespace-schema-defect",
		},
		{
			ID: "known-mixin-annotation-retention", Severity: "critical", Category: "mixin", KnownRepair: true,
			Title:     "Known Mixin annotation contract failure matches a superseded repair attempt",
			Action:    "Inspect RuntimeVisibleAnnotations and RuntimeInvisibleAnnotations with javap -verbose. Compile against the real Mixin API or reproduce CLASS retention and exact targets/descriptors before rebuilding affected mixins.",
			Needles:   []string{"does not contain a mixin annotation", "mixin annotation", "invalidmixinexception"},
			SourceIDs: []string{"repair-brain", "mixin-source", "mixin-wiki", "mixintrace"}, RepairPatternID: "mixin-annotation-contract",
		},
		{
			ID: "unsupported-class-version", Severity: "critical", Category: "java",
			Title:     "The runtime Java version cannot load one or more class files",
			Action:    "Identify the named class/JAR and class-file major version. Use the Java version required by the Minecraft release or rebuild the mod for the target bytecode level; metadata-only edits cannot fix class format.",
			Needles:   []string{"unsupportedclassversionerror", "class file version", "only recognizes class file versions up to"},
			SourceIDs: []string{"jdk-tools", "gradle-toolchains"},
		},
		{
			ID: "missing-class", Severity: "critical", Category: "dependency",
			Title:     "A required class is absent at runtime",
			Action:    "Trace the earliest ClassNotFoundException or NoClassDefFoundError to its owning mod/library. Resolve missing, wrong-version, wrong-loader, shaded, or side-only dependencies before addressing later cascades.",
			Needles:   []string{"classnotfoundexception", "noclassdeffounderror", "could not find required mod", "missing mandatory dependencies"},
			SourceIDs: []string{"classgraph", "jdk-tools", "modrinth-api", "curseforge-api"}, RepairPatternID: "loader-metadata-not-proof",
		},
		{
			ID: "method-linkage", Severity: "critical", Category: "binary-api",
			Title:     "A linked method descriptor changed or disappeared",
			Action:    "Prove the exact owner, method name, descriptor, dependent callsite, and installed provider bytecode. Update the dependent mod, patch the narrow callsite, or add a version-gated adapter only after equivalent behavior is proven.",
			Needles:   []string{"nosuchmethoderror", "abstractmethoderror", "incompatibleclasschangeerror"},
			SourceIDs: []string{"japicmp", "revapi", "asm", "recaf", "vineflower"}, RepairPatternID: "binary-owner-descriptor-drift",
		},
		{
			ID: "field-linkage", Severity: "critical", Category: "binary-api",
			Title:     "A linked field changed or disappeared",
			Action:    "Compare the dependent constant-pool field reference to the installed target class. Prove the replacement field or accessor and patch/recompile only the affected dependent code.",
			Needles:   []string{"nosuchfielderror"},
			SourceIDs: []string{"japicmp", "revapi", "asm", "recaf", "vineflower"}, RepairPatternID: "binary-owner-descriptor-drift",
		},
		{
			ID: "mixin-application", Severity: "critical", Category: "mixin",
			Title:     "Mixin transformation failed before or during application",
			Action:    "Identify the exact mixin, target class, member descriptor, injection point, ordinal/slice, priority, refmap, and competing transformer. Compare against the exact target bytecode and retarget the smallest broken path.",
			Needles:   []string{"mixinapplyerror", "mixintransformererror", "injectionerror", "critical injection failure", "invalidinjectorexception", "invalidmixinexception"},
			SourceIDs: []string{"mixin-source", "mixin-wiki", "mixintrace", "vineflower", "enigma"}, RepairPatternID: "mixin-annotation-contract",
		},
		{
			ID: "duplicate-mods", Severity: "critical", Category: "dependency",
			Title:     "Duplicate mods or duplicate mod identifiers were detected",
			Action:    "Hash and inspect every named file, keep the intended loader/game build, and remove only exact or proven superseded duplicates. Re-run dependency resolution before launch.",
			Needles:   []string{"duplicate mods found", "duplicate mod", "duplicate mod id", "found duplicate mods"},
			SourceIDs: []string{"modrinth-api", "curseforge-api", "packwiz", "ferium"},
		},
		{
			ID: "wrong-side-class", Severity: "critical", Category: "sides",
			Title:     "Client-only code loaded on a dedicated server or vice versa",
			Action:    "Trace the first side-only class reference and move registration/initialization behind the loader's physical-side boundary. Verify both client and dedicated-server launches.",
			Needles:   []string{"invalid dist dedicated_server", "attempted to load class net/minecraft/client", "cannot load class net.minecraft.client", "wrong side"},
			SourceIDs: []string{"neoforge-gametest", "fabric-testing", "headlessmc"},
		},
		{
			ID: "registry-or-data-bootstrap", Severity: "critical", Category: "data",
			Title:     "Registry, codec, datapack, or worldgen bootstrap failed",
			Action:    "Follow the earliest missing key, codec, tag, pool, recipe, loot, or registry-remap error. Validate packaged data against the target schema and test on a copied world before changing persistent data.",
			Needles:   []string{"unknown registry key", "unbound values", "missing referenced structure", "failed to parse datapack", "registry remap failed", "codec error"},
			SourceIDs: []string{"minecraft-wiki", "spyglass", "nbt-studio", "datafixerupper", "amulet"}, RepairPatternID: "data-namespace-schema-defect",
		},
		{
			ID: "out-of-memory", Severity: "high", Category: "resources",
			Title:     "The JVM exhausted heap or native memory",
			Action:    "Establish correctness first, then capture heap/native diagnostics and a reproducible workload. Distinguish leaks, runaway generation, oversized assets, native allocation, and an undersized JVM before tuning memory.",
			Needles:   []string{"outofmemoryerror", "java heap space", "gc overhead limit exceeded", "unable to create native thread", "direct buffer memory"},
			SourceIDs: []string{"spark", "jdk-tools"},
		},
		{
			ID: "graphics-or-native", Severity: "high", Category: "rendering",
			Title:     "Graphics or native-library initialization failed",
			Action:    "Identify the first native/graphics exception, exact driver/runtime, renderer/shader stack, and extracted native path. Reproduce without unrelated warning noise and fix the actual ABI or renderer conflict without reducing fidelity.",
			Needles:   []string{"org.lwjgl", "glfw error", "unsatisfiedlinkerror", "failed to load native", "vulkan initialization", "opengl error"},
			SourceIDs: []string{"mclogs", "spark", "jdk-tools"},
		},
	}
}

func doctorSelectRootCause(findings []DoctorFinding) DoctorFinding {
	if len(findings) == 0 {
		return DoctorFinding{}
	}
	rank := map[string]int{"critical": 4, "high": 3, "warning": 2, "info": 1}
	best := findings[0]
	for _, finding := range findings[1:] {
		if rank[finding.Severity] > rank[best.Severity] {
			best = finding
		}
	}
	return best
}

func doctorContainsAny(lower string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func doctorMatchingLogLines(lines, needles []string, limit int) []string {
	out := []string{}
	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, needle := range needles {
			if needle != "" && strings.Contains(lower, strings.ToLower(needle)) {
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					out = appendUniqueString(out, trimmed)
				}
				break
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func doctorFirstNonEmptyLogLine(lines []string) string {
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return "No non-empty log lines supplied"
}

func sortDoctorPatterns(patterns []DoctorRepairPattern) {
	sort.SliceStable(patterns, func(i, j int) bool {
		return strings.ToLower(patterns[i].Name) < strings.ToLower(patterns[j].Name)
	})
}
