package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreatorOCRReadsMinecraftModNamesWhenLocalToolsExist(t *testing.T) {
	tesseract, err := exec.LookPath("tesseract")
	if err != nil {
		t.Skip("tesseract not installed")
	}
	magick, err := exec.LookPath("magick")
	if err != nil {
		t.Skip("ImageMagick not installed")
	}
	frame := filepath.Join(t.TempDir(), "frame.png")
	cmd := exec.Command(magick, "-size", "1200x500", "xc:white", "-fill", "black", "-font", "DejaVu-Sans", "-pointsize", "72", "-gravity", "center", "-annotate", "+0+0", "Better Combat\nCombat Roll", frame)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("could not create OCR fixture: %v (%s)", err, strings.TrimSpace(string(out)))
	}
	app := &App{cfgDir: t.TempDir()}
	text, err := app.ocrCreatorFrame(context.Background(), creatorOCRBackend{Name: "Tesseract OCR", Command: tesseract, Kind: "tesseract"}, frame)
	if err != nil {
		t.Fatalf("OCR failed: %v", err)
	}
	low := strings.ToLower(text)
	if !strings.Contains(low, "better combat") || !strings.Contains(low, "combat roll") {
		t.Fatalf("OCR text=%q did not preserve expected mod names", text)
	}
}
