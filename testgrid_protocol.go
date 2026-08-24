package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var bedrockRakNetMagic = []byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

type BedrockStatus struct {
	Edition     string   `json:"edition,omitempty"`
	MOTD        string   `json:"motd,omitempty"`
	Protocol    int      `json:"protocol,omitempty"`
	Version     string   `json:"version,omitempty"`
	Players     int      `json:"players,omitempty"`
	MaxPlayers  int      `json:"maxPlayers,omitempty"`
	ServerID    string   `json:"serverId,omitempty"`
	SubMOTD     string   `json:"subMotd,omitempty"`
	GameMode    string   `json:"gameMode,omitempty"`
	GameModeID  int      `json:"gameModeId,omitempty"`
	IPv4Port    int      `json:"ipv4Port,omitempty"`
	IPv6Port    int      `json:"ipv6Port,omitempty"`
	Raw         string   `json:"raw"`
	ExtraFields []string `json:"extraFields,omitempty"`
}

func waitForTCP(ctx context.Context, address string, intervalMS int) error {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return fmt.Errorf("invalid address %q: %w", address, err)
	}
	ticker := time.NewTicker(time.Duration(intervalMS) * time.Millisecond)
	defer ticker.Stop()
	for {
		dialer := net.Dialer{Timeout: time.Second}
		connection, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = connection.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("TCP endpoint %s not ready: %w", address, ctx.Err())
		case <-ticker.C:
		}
	}
}

func waitForJavaStatus(ctx context.Context, address string, intervalMS int) (map[string]any, error) {
	ticker := time.NewTicker(time.Duration(intervalMS) * time.Millisecond)
	defer ticker.Stop()
	var lastErr error
	for {
		status, err := minecraftJavaStatus(ctx, address)
		if err == nil {
			return status, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Java status endpoint %s not ready (%v): %w", address, lastErr, ctx.Err())
		case <-ticker.C:
		}
	}
}

func minecraftJavaStatus(ctx context.Context, address string) (map[string]any, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("invalid Java server port")
	}
	dialer := net.Dialer{Timeout: 3 * time.Second}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(4 * time.Second))

	handshake := &bytes.Buffer{}
	writeMinecraftVarInt(handshake, 0)
	writeMinecraftVarInt(handshake, -1)
	writeMinecraftString(handshake, host)
	_ = binary.Write(handshake, binary.BigEndian, uint16(port))
	writeMinecraftVarInt(handshake, 1)
	if err := writeMinecraftPacket(connection, handshake.Bytes()); err != nil {
		return nil, err
	}
	if err := writeMinecraftPacket(connection, []byte{0}); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(connection)
	length, err := readMinecraftVarInt(reader)
	if err != nil {
		return nil, err
	}
	if length < 1 || length > 2<<20 {
		return nil, fmt.Errorf("invalid Java status packet length %d", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	payloadReader := bytes.NewReader(payload)
	packetID, err := readMinecraftVarInt(payloadReader)
	if err != nil || packetID != 0 {
		return nil, fmt.Errorf("unexpected Java status packet id %d", packetID)
	}
	jsonLength, err := readMinecraftVarInt(payloadReader)
	if err != nil || jsonLength < 0 || jsonLength > payloadReader.Len() {
		return nil, errors.New("invalid Java status JSON length")
	}
	body := make([]byte, jsonLength)
	if _, err := io.ReadFull(payloadReader, body); err != nil {
		return nil, err
	}
	var status map[string]any
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("decode Java status: %w", err)
	}
	status["address"] = address
	return status, nil
}

func writeMinecraftPacket(writer io.Writer, payload []byte) error {
	var prefix bytes.Buffer
	writeMinecraftVarInt(&prefix, int32(len(payload)))
	if _, err := writer.Write(prefix.Bytes()); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func writeMinecraftString(writer io.Writer, value string) {
	writeMinecraftVarInt(writer, int32(len(value)))
	_, _ = io.WriteString(writer, value)
}

func writeMinecraftVarInt(writer io.Writer, value int32) {
	unsigned := uint32(value)
	for {
		current := byte(unsigned & 0x7f)
		unsigned >>= 7
		if unsigned != 0 {
			current |= 0x80
		}
		_, _ = writer.Write([]byte{current})
		if unsigned == 0 {
			return
		}
	}
}

func readMinecraftVarInt(reader io.ByteReader) (int, error) {
	var result uint32
	for position := uint(0); position < 35; position += 7 {
		current, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint32(current&0x7f) << position
		if current&0x80 == 0 {
			return int(int32(result)), nil
		}
	}
	return 0, errors.New("Minecraft VarInt exceeds five bytes")
}

func waitForBedrockStatus(ctx context.Context, address string, intervalMS int) (BedrockStatus, error) {
	ticker := time.NewTicker(time.Duration(intervalMS) * time.Millisecond)
	defer ticker.Stop()
	var lastErr error
	for {
		status, err := minecraftBedrockStatus(ctx, address)
		if err == nil {
			return status, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return BedrockStatus{}, fmt.Errorf("Bedrock status endpoint %s not ready (%v): %w", address, lastErr, ctx.Err())
		case <-ticker.C:
		}
	}
}

func minecraftBedrockStatus(ctx context.Context, address string) (BedrockStatus, error) {
	endpoint, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return BedrockStatus{}, err
	}
	dialer := net.Dialer{Timeout: 3 * time.Second}
	connection, err := dialer.DialContext(ctx, "udp", endpoint.String())
	if err != nil {
		return BedrockStatus{}, err
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(4 * time.Second))
	request := &bytes.Buffer{}
	request.WriteByte(0x01)
	_ = binary.Write(request, binary.BigEndian, time.Now().UnixMilli())
	request.Write(bedrockRakNetMagic)
	_ = binary.Write(request, binary.BigEndian, rand.Int63())
	if _, err := connection.Write(request.Bytes()); err != nil {
		return BedrockStatus{}, err
	}
	response := make([]byte, 65535)
	n, err := connection.Read(response)
	if err != nil {
		return BedrockStatus{}, err
	}
	response = response[:n]
	if len(response) < 35 || response[0] != 0x1c {
		return BedrockStatus{}, errors.New("invalid Bedrock unconnected pong")
	}
	if !bytes.Equal(response[17:33], bedrockRakNetMagic) {
		return BedrockStatus{}, errors.New("invalid Bedrock RakNet magic")
	}
	length := int(binary.BigEndian.Uint16(response[33:35]))
	if length < 0 || 35+length > len(response) {
		return BedrockStatus{}, errors.New("invalid Bedrock status string length")
	}
	raw := string(response[35 : 35+length])
	fields := strings.Split(raw, ";")
	status := BedrockStatus{Raw: raw}
	if len(fields) > 0 {
		status.Edition = fields[0]
	}
	if len(fields) > 1 {
		status.MOTD = fields[1]
	}
	if len(fields) > 2 {
		status.Protocol, _ = strconv.Atoi(fields[2])
	}
	if len(fields) > 3 {
		status.Version = fields[3]
	}
	if len(fields) > 4 {
		status.Players, _ = strconv.Atoi(fields[4])
	}
	if len(fields) > 5 {
		status.MaxPlayers, _ = strconv.Atoi(fields[5])
	}
	if len(fields) > 6 {
		status.ServerID = fields[6]
	}
	if len(fields) > 7 {
		status.SubMOTD = fields[7]
	}
	if len(fields) > 8 {
		status.GameMode = fields[8]
	}
	if len(fields) > 9 {
		status.GameModeID, _ = strconv.Atoi(fields[9])
	}
	if len(fields) > 10 {
		status.IPv4Port, _ = strconv.Atoi(fields[10])
	}
	if len(fields) > 11 {
		status.IPv6Port, _ = strconv.Atoi(fields[11])
	}
	if len(fields) > 12 {
		status.ExtraFields = fields[12:]
	}
	return status, nil
}

func minecraftRCON(ctx context.Context, address, password, command string) (string, error) {
	if address == "" || password == "" || command == "" {
		return "", errors.New("RCON address, password, and command are required")
	}
	dialer := net.Dialer{Timeout: 3 * time.Second}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return "", err
	}
	defer connection.Close()
	deadline := time.Now().Add(5 * time.Second)
	if value, ok := ctx.Deadline(); ok && value.Before(deadline) {
		deadline = value
	}
	_ = connection.SetDeadline(deadline)
	requestID := int32(0x4d4d5601)
	if err := writeRCONPacket(connection, requestID, 3, password); err != nil {
		return "", err
	}
	authID, _, _, err := readRCONPacket(connection)
	if err != nil {
		return "", err
	}
	if authID == -1 {
		return "", errors.New("RCON authentication failed")
	}
	if err := writeRCONPacket(connection, requestID+1, 2, command); err != nil {
		return "", err
	}
	responseID, _, response, err := readRCONPacket(connection)
	if err != nil {
		return "", err
	}
	if responseID != requestID+1 {
		return "", errors.New("RCON response id mismatch")
	}
	return response, nil
}

func writeRCONPacket(writer io.Writer, requestID, packetType int32, body string) error {
	payload := &bytes.Buffer{}
	_ = binary.Write(payload, binary.LittleEndian, requestID)
	_ = binary.Write(payload, binary.LittleEndian, packetType)
	payload.WriteString(body)
	payload.Write([]byte{0, 0})
	if err := binary.Write(writer, binary.LittleEndian, int32(payload.Len())); err != nil {
		return err
	}
	_, err := writer.Write(payload.Bytes())
	return err
}

func readRCONPacket(reader io.Reader) (int32, int32, string, error) {
	var length int32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return 0, 0, "", err
	}
	if length < 10 || length > 4<<20 {
		return 0, 0, "", fmt.Errorf("invalid RCON packet length %d", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return 0, 0, "", err
	}
	requestID := int32(binary.LittleEndian.Uint32(payload[0:4]))
	packetType := int32(binary.LittleEndian.Uint32(payload[4:8]))
	body := strings.TrimRight(string(payload[8:]), "\x00")
	return requestID, packetType, body, nil
}
