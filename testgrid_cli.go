package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func runTestGridCLI(arguments []string) int {
	if len(arguments) == 0 || arguments[0] == "help" || arguments[0] == "--help" || arguments[0] == "-h" {
		printTestGridUsage()
		return 0
	}
	command := strings.ToLower(arguments[0])
	switch command {
	case "capabilities":
		return printTestGridJSON(testGridCapabilities())
	case "validate":
		if len(arguments) != 2 {
			fmt.Fprintln(os.Stderr, "usage: MinecraftModVault testgrid validate <manifest.json>")
			return 2
		}
		manifest, err := loadTestGridManifest(arguments[1])
		if err == nil {
			manifest = normalizeTestGridManifest(manifest)
			err = validateTestGridManifest(manifest)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		return printTestGridJSON(map[string]any{"valid": true, "manifest": arguments[1], "schemaVersion": manifest.SchemaVersion})
	case "run":
		if len(arguments) != 2 {
			fmt.Fprintln(os.Stderr, "usage: MinecraftModVault testgrid run <manifest.json>")
			return 2
		}
		manifest, err := loadTestGridManifest(arguments[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		grid, err := commandLineTestGrid()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		run, runErr := grid.Run(ctx, manifest)
		_ = printTestGridJSON(run)
		if runErr != nil {
			return 1
		}
		return 0
	case "inspect":
		if len(arguments) != 2 {
			fmt.Fprintln(os.Stderr, "usage: MinecraftModVault testgrid inspect <run-id>")
			return 2
		}
		grid, err := commandLineTestGrid()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		run, ok := grid.Get(arguments[1])
		if !ok {
			fmt.Fprintln(os.Stderr, "TestGrid run not found")
			return 1
		}
		return printTestGridJSON(run)
	case "ping-java":
		if len(arguments) != 2 {
			fmt.Fprintln(os.Stderr, "usage: MinecraftModVault testgrid ping-java <host:port>")
			return 2
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		status, err := minecraftJavaStatus(ctx, arguments[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return printTestGridJSON(status)
	case "ping-bedrock":
		if len(arguments) != 2 {
			fmt.Fprintln(os.Stderr, "usage: MinecraftModVault testgrid ping-bedrock <host:port>")
			return 2
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		status, err := minecraftBedrockStatus(ctx, arguments[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return printTestGridJSON(status)
	default:
		fmt.Fprintf(os.Stderr, "unknown TestGrid command %q\n", arguments[0])
		printTestGridUsage()
		return 2
	}
}

func configureServeCommand(arguments []string) error {
	_ = os.Setenv("MMV_NO_LAUNCH", "1")
	_ = os.Setenv("MMV_HEADLESS", "1")
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
		case "--port":
			index++
			if index >= len(arguments) {
				return errors.New("--port requires a value")
			}
			port, err := strconv.Atoi(arguments[index])
			if err != nil || port < 0 || port > 65535 {
				return errors.New("--port must be between 0 and 65535")
			}
			_ = os.Setenv("MMV_PORT", strconv.Itoa(port))
		case "--token-file":
			index++
			if index >= len(arguments) {
				return errors.New("--token-file requires a path")
			}
			_ = os.Setenv("MMV_TOKEN_FILE", arguments[index])
		case "--config-dir":
			index++
			if index >= len(arguments) {
				return errors.New("--config-dir requires a path")
			}
			_ = os.Setenv("MMV_CONFIG_DIR", arguments[index])
		case "--help", "-h":
			fmt.Println("usage: MinecraftModVault serve [--port 0-65535] [--token-file path] [--config-dir path]")
			return errServeHelp
		default:
			return fmt.Errorf("unknown serve option %q", arguments[index])
		}
	}
	return nil
}

var errServeHelp = errors.New("serve help requested")

func commandLineTestGrid() (*TestGrid, error) {
	cfg, err := configDir()
	if err != nil {
		return nil, err
	}
	return newTestGrid(filepath.Join(cfg, "testgrid"))
}

func loadTestGridManifest(path string) (TestGridManifest, error) {
	var manifest TestGridManifest
	body, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return manifest, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func printTestGridJSON(value any) int {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func printTestGridUsage() {
	fmt.Fprintln(os.Stderr, `Minecraft Mod Vault TestGrid

Usage:
  MinecraftModVault testgrid capabilities
  MinecraftModVault testgrid validate <manifest.json>
  MinecraftModVault testgrid run <manifest.json>
  MinecraftModVault testgrid inspect <run-id>
  MinecraftModVault testgrid ping-java <host:port>
  MinecraftModVault testgrid ping-bedrock <host:port>`)
}
