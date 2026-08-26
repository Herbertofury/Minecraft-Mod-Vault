package main

import (
	"fmt"
	"os"
)

const version = "2.2.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
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
	case "adopt-tools":
		runAdoptTools(os.Args[2:])
	case "heritage":
		runHeritage(os.Args[2:])
	case "port-guard":
		runPortGuard(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}
func usage() {
	fmt.Print(`Minecraft Dev Kit Orchestrator 2.2.0

Commands:
  mmv-devkit validate --registry devkit-registry.json
  mmv-devkit check --registry devkit-registry.json [--mc 1.21.1] [--loader neoforge] [--json]
  mmv-devkit sync --registry devkit-registry.json [--mc ...] [--loader ...] [--apply] [--drive]
  mmv-devkit watch --registry devkit-registry.json --drive [--interval 15m]
  mmv-devkit sources --catalog tool-provenance.json [--id geckolib-forge] [--json]
  mmv-devkit heritage --registry devkit-registry.json --id MOD --mc 1.20.1 --loader forge --out .mmv/heritage --report heritage.json
  mmv-devkit port-guard --registry devkit-registry.json --id MOD --original original.jar --converted build.jar --converted-source src/main --mc 1.20.1 --loader forge --report port-guard.json

Credentials are read from environment only:
  CURSEFORGE_API_KEY, GITHUB_TOKEN, MMV_GOOGLE_DRIVE_TOKEN
`)
}
