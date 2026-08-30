package main

import (
	"fmt"
	"os"
)

const version = "2.7.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "help", "-h", "--help":
		usage()
		return
	case "version":
		fmt.Println("Minecraft Dev Kit Orchestrator", version)
	case "validate":
		runValidate(os.Args[2:])
	case "check":
		runCheck(os.Args[2:])
	case "sync":
		runSync(os.Args[2:])
	case "watch":
		runWatch(os.Args[2:])
	case "sources":
		runSources(os.Args[2:])
	case "fetch-mod":
		runFetchMod(os.Args[2:])
	case "adopt-tools":
		runAdoptTools(os.Args[2:])
	case "heritage":
		runHeritage(os.Args[2:])
	case "port-guard":
		runPortGuard(os.Args[2:])
	case "cache-doctor":
		runCacheDoctor(os.Args[2:])
	case "cache-reassemble":
		runCacheReassemble(os.Args[2:])
	case "archive-split":
		runArchiveSplit(os.Args[2:])
	case "client-assets":
		runClientAssets(os.Args[2:])
	case "client-natives":
		runClientNatives(os.Args[2:])
	case "world-qa-enable":
		runWorldQAEnable(os.Args[2:])
	case "compat":
		runCompat(os.Args[2:])
	case "vanilla-atlas":
		runVanillaAtlas(os.Args[2:])
	case "multiversion":
		runMultiVersion(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Printf(`Minecraft Dev Kit Orchestrator %s

Commands:
  mmv-devkit validate --registry devkit-registry.json
  mmv-devkit check --registry devkit-registry.json [--mc 1.21.1] [--loader neoforge] [--json]
  mmv-devkit sync --registry devkit-registry.json [--mc ...] [--loader ...] [--apply] [--drive]
  mmv-devkit watch --registry devkit-registry.json --drive [--interval 15m]
  mmv-devkit sources --catalog tool-provenance.json [--id geckolib-forge] [--json]
  mmv-devkit fetch-mod --provider modrinth --project PROJECT --version VERSION [--out FILE] [--receipt fetch.json] [--json]
  mmv-devkit fetch-mod --provider curseforge --project MOD_ID --file FILE_ID [--out FILE] [--receipt fetch.json] [--json]
  mmv-devkit adopt-tools --registry devkit-registry.json --catalog tool-provenance.json --drive
  mmv-devkit heritage --registry devkit-registry.json --id MOD --mc 1.20.1 --loader forge --out .mmv/heritage --report heritage.json
  mmv-devkit port-guard --registry devkit-registry.json --id MOD --original original.jar --converted build.jar --converted-source src/main --mc 1.20.1 --loader forge --report port-guard.json
  mmv-devkit cache-doctor --cache GRADLE_HOME --mc 1.20.1 --forge 47.4.23 --forgegradle 6.0.54 [--expect-min-files 1600] [--archive CACHE.zip] [--json]
  mmv-devkit archive-split --file BIG.zip [--out-dir BIG.zip-SPLIT] [--part-mib 85] [--json]
  mmv-devkit cache-reassemble --manifest SPLIT-MANIFEST.json [--parts-dir DIR] [--out CACHE.zip] [--extract GRADLE_HOME] [--json]
  mmv-devkit client-assets --gradle-home GRADLE_HOME --mc 1.20.1 [--assets-dir DIR] [--workers 16] [--verify-only] [--json]
  mmv-devkit client-natives --gradle-home GRADLE_HOME --platform linux --out PROJECT/build/mmv-natives [--json]
  mmv-devkit world-qa-enable --level-dat PROJECT/run/saves/Creeperella-QA/level.dat [--verify-only] [--json]
  mmv-devkit compat init|ingest|matrix|plan|scaffold|verify ...
  mmv-devkit vanilla-atlas verify|query|sound-status|diff|plan-backport|providers|build-backport-catalog|resolve-owner|validate-capabilities ...
  mmv-devkit multiversion scaffold|validate ...

Credentials are read from environment only:
  CURSEFORGE_API_KEY (aliases: CF_API_KEY, CURSEFORGE_KEY), GITHUB_TOKEN, MMV_GOOGLE_DRIVE_TOKEN
`, version)
}
