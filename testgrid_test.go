package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGridHelperProcess(t *testing.T) {
	if os.Getenv("MMV_TESTGRID_HELPER") != "1" {
		return
	}
	fmt.Println("TESTGRID_READY version=fixture")
	if err := os.WriteFile("evidence.txt", []byte("deterministic evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTestGridRunsProcessAssertionsAndArtifacts(t *testing.T) {
	workingDirectory := t.TempDir()
	grid, err := newTestGrid(filepath.Join(t.TempDir(), "testgrid"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := TestGridManifest{
		SchemaVersion:  1,
		Name:           "fixture lifecycle",
		Edition:        "generic",
		TimeoutSeconds: 20,
		Runtime: TestGridRuntime{
			Kind:             "build-or-tool",
			Executable:       os.Args[0],
			Arguments:        []string{"-test.run=^TestGridHelperProcess$"},
			WorkingDirectory: workingDirectory,
			Environment:      map[string]string{"MMV_TESTGRID_HELPER": "1"},
		},
		Steps: []TestGridStep{
			{Name: "ready marker", Type: "wait-log", Pattern: "TESTGRID_READY", TimeoutSeconds: 5},
			{Name: "no panic", Type: "deny-log", Pattern: "panic:", TimeoutSeconds: 5},
			{Name: "evidence exists", Type: "file-exists", Path: "evidence.txt", TimeoutSeconds: 5},
		},
		Artifacts: []TestGridArtifact{{Name: "fixture-evidence.txt", Path: "evidence.txt"}},
	}
	run, err := grid.Run(context.Background(), manifest)
	if err != nil {
		t.Fatalf("run failed: %v\n%+v", err, run)
	}
	if run.Status != "passed" || run.Process.ExitCode != 0 {
		t.Fatalf("unexpected run result: %+v", run)
	}
	if len(run.Steps) != 3 || len(run.Artifacts) != 1 {
		t.Fatalf("missing evidence: steps=%d artifacts=%d", len(run.Steps), len(run.Artifacts))
	}
	for _, path := range []string{run.ReportPath, run.JUnitPath, run.HTMLPath, run.LogPath, run.Artifacts[0].Path} {
		if info, statErr := os.Stat(path); statErr != nil || info.IsDir() {
			t.Fatalf("expected report artifact %s: %v", path, statErr)
		}
	}
	reloaded, ok := grid.Get(run.ID)
	if !ok || reloaded.ID != run.ID || reloaded.Status != "passed" {
		t.Fatalf("persisted run not readable: ok=%v run=%+v", ok, reloaded)
	}
}

func TestTestGridRejectsUnknownStepAndRedactsSecrets(t *testing.T) {
	manifest := normalizeTestGridManifest(TestGridManifest{
		Name:    "bad step",
		Edition: "java",
		Runtime: TestGridRuntime{Kind: "custom", Executable: os.Args[0], Environment: map[string]string{"RCON_PASSWORD": "secret", "NORMAL": "visible"}},
		Steps:   []TestGridStep{{Type: "telepathy"}},
	})
	if err := validateTestGridManifest(manifest); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported-step error, got %v", err)
	}
	redacted := redactTestGridManifest(manifest)
	if redacted.Runtime.Environment["RCON_PASSWORD"] != "[redacted]" || redacted.Runtime.Environment["NORMAL"] != "visible" {
		t.Fatalf("unexpected redaction: %#v", redacted.Runtime.Environment)
	}
}

func TestMinecraftJavaStatusProtocol(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverError := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverError <- acceptErr
			return
		}
		defer connection.Close()
		_ = connection.SetDeadline(time.Now().Add(3 * time.Second))
		reader := bufio.NewReader(connection)
		for packet := 0; packet < 2; packet++ {
			length, readErr := readMinecraftVarInt(reader)
			if readErr != nil {
				serverError <- readErr
				return
			}
			request := make([]byte, length)
			if _, readErr := io.ReadFull(reader, request); readErr != nil {
				serverError <- readErr
				return
			}
		}
		statusJSON := `{"version":{"name":"1.20.1","protocol":763},"players":{"max":8,"online":1},"description":{"text":"Vault fixture"}}`
		payload := &bytes.Buffer{}
		writeMinecraftVarInt(payload, 0)
		writeMinecraftString(payload, statusJSON)
		serverError <- writeMinecraftPacket(connection, payload.Bytes())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status, err := minecraftJavaStatus(ctx, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	version, _ := status["version"].(map[string]any)
	if version["name"] != "1.20.1" || status["address"] != listener.Addr().String() {
		t.Fatalf("unexpected Java status: %#v", status)
	}
	if err := <-serverError; err != nil {
		t.Fatal(err)
	}
}

func TestMinecraftBedrockStatusProtocol(t *testing.T) {
	server, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	serverError := make(chan error, 1)
	go func() {
		request := make([]byte, 2048)
		_, client, readErr := server.ReadFromUDP(request)
		if readErr != nil {
			serverError <- readErr
			return
		}
		status := "MCPE;Vault Fixture;818;1.21.90;2;16;1234;Sub MOTD;Survival;1;19132;19133;false"
		response := &bytes.Buffer{}
		response.WriteByte(0x1c)
		_ = binary.Write(response, binary.BigEndian, time.Now().UnixMilli())
		_ = binary.Write(response, binary.BigEndian, int64(1234))
		response.Write(bedrockRakNetMagic)
		_ = binary.Write(response, binary.BigEndian, uint16(len(status)))
		response.WriteString(status)
		_, writeErr := server.WriteToUDP(response.Bytes(), client)
		serverError <- writeErr
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status, err := minecraftBedrockStatus(ctx, server.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	if status.Edition != "MCPE" || status.Version != "1.21.90" || status.Players != 2 || status.MaxPlayers != 16 || status.GameMode != "Survival" {
		t.Fatalf("unexpected Bedrock status: %+v", status)
	}
	if err := <-serverError; err != nil {
		t.Fatal(err)
	}
}

func TestMinecraftRCONProtocol(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverError := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverError <- acceptErr
			return
		}
		defer connection.Close()
		authID, packetType, password, readErr := readRCONPacket(connection)
		if readErr != nil {
			serverError <- readErr
			return
		}
		if packetType != 3 || password != "fixture-password" {
			serverError <- fmt.Errorf("unexpected auth packet type=%d password=%q", packetType, password)
			return
		}
		if writeErr := writeRCONPacket(connection, authID, 2, ""); writeErr != nil {
			serverError <- writeErr
			return
		}
		commandID, packetType, command, readErr := readRCONPacket(connection)
		if readErr != nil {
			serverError <- readErr
			return
		}
		if packetType != 2 || command != "list" {
			serverError <- fmt.Errorf("unexpected command packet type=%d command=%q", packetType, command)
			return
		}
		serverError <- writeRCONPacket(connection, commandID, 0, "There are 0 of a max of 20 players online")
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := minecraftRCON(ctx, listener.Addr().String(), "fixture-password", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(response, "max of 20") {
		t.Fatalf("unexpected RCON response %q", response)
	}
	if err := <-serverError; err != nil {
		t.Fatal(err)
	}
}

func TestTestGridHTTPRoutesRequireTokenAndValidate(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), token: "fixture-token"}
	mux := http.NewServeMux()
	app.registerRoutes(mux)

	unauthorized := httptest.NewRecorder()
	mux.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/testgrid/capabilities", nil))
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status=%d", unauthorized.Code)
	}

	authorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/testgrid/capabilities", nil)
	request.Header.Set("X-MMV-Token", "fixture-token")
	mux.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK || !strings.Contains(authorized.Body.String(), "bedrock-ping") {
		t.Fatalf("capabilities status=%d body=%s", authorized.Code, authorized.Body.String())
	}

	badManifest, _ := json.Marshal(TestGridManifest{Name: "invalid", Edition: "java", Runtime: TestGridRuntime{Executable: os.Args[0]}, Steps: []TestGridStep{{Type: "unknown"}}})
	bad := httptest.NewRecorder()
	badRequest := httptest.NewRequest(http.MethodPost, "/api/testgrid/run", bytes.NewReader(badManifest))
	badRequest.Header.Set("X-MMV-Token", "fixture-token")
	mux.ServeHTTP(bad, badRequest)
	if bad.Code != http.StatusBadRequest {
		body, _ := io.ReadAll(bad.Result().Body)
		t.Fatalf("bad manifest status=%d body=%s", bad.Code, body)
	}
}
