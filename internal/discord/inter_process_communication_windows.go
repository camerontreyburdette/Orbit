//go:build windows

package discord

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	operationCodeHandshake  = 0
	operationCodeFrame      = 1
	frameHeaderLength       = 8
	maximumResponseLength   = 65536
	pipeCandidateCount      = 10
	handshakeResponseWait   = 600 * time.Millisecond
	activityResponseWait    = 600 * time.Millisecond
	clearActivityWait       = 200 * time.Millisecond
	pipePollInterval        = 10 * time.Millisecond
	discordPipePathTemplate = `\\.\pipe\discord-ipc-%d`
)

var (
	kernelLibrary          = syscall.NewLazyDLL("kernel32.dll")
	peekNamedPipeProcedure = kernelLibrary.NewProc("PeekNamedPipe")
)

type pipeConnection struct {
	mutex                   sync.Mutex
	handle                  windows.Handle
	isConnected             bool
	currentClientIdentifier string
	lastActivityJsonPayload string
}

func newPipeConnection() *pipeConnection {
	return &pipeConnection{}
}

func peekAvailableBytes(handle windows.Handle) (uint32, error) {
	var totalBytesAvailable uint32
	result, _, systemError := peekNamedPipeProcedure.Call(
		uintptr(handle),
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&totalBytesAvailable)),
		0,
	)
	if result == 0 {
		return 0, systemError
	}
	return totalBytesAvailable, nil
}

func readWithTimeout(handle windows.Handle, buffer []byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		bytesAvailable, peekError := peekAvailableBytes(handle)
		if peekError != nil {
			return peekError
		}
		if bytesAvailable >= uint32(len(buffer)) {
			var bytesRead uint32
			return windows.ReadFile(handle, buffer, &bytesRead, nil)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("read timeout on discord pipe")
		}
		time.Sleep(pipePollInterval)
	}
}

func writeFrame(handle windows.Handle, operationCode uint32, payload []byte) error {
	frame := make([]byte, frameHeaderLength+len(payload))
	binary.LittleEndian.PutUint32(frame[0:4], operationCode)
	binary.LittleEndian.PutUint32(frame[4:8], uint32(len(payload)))
	copy(frame[frameHeaderLength:], payload)

	var writtenBytes uint32
	return windows.WriteFile(handle, frame, &writtenBytes, nil)
}

func readFrameLength(handle windows.Handle, timeout time.Duration) (uint32, error) {
	responseHeader := make([]byte, frameHeaderLength)
	if readHeaderError := readWithTimeout(handle, responseHeader, timeout); readHeaderError != nil {
		return 0, readHeaderError
	}
	responseLength := binary.LittleEndian.Uint32(responseHeader[4:8])
	if responseLength > maximumResponseLength {
		return 0, fmt.Errorf("response length too large: %d", responseLength)
	}
	return responseLength, nil
}

func readFrameBody(handle windows.Handle, responseLength uint32, timeout time.Duration) error {
	responseData := make([]byte, responseLength)
	return readWithTimeout(handle, responseData, timeout)
}

func activityPayload(activity interface{}) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": activity,
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	})
}

func (connection *pipeConnection) isAlive() bool {
	if connection.handle == 0 || !connection.isConnected {
		return false
	}
	_, peekError := peekAvailableBytes(connection.handle)
	return peekError == nil
}

func (connection *pipeConnection) resetLocked() {
	if connection.handle != 0 {
		_ = windows.CloseHandle(connection.handle)
	}
	connection.handle = 0
	connection.isConnected = false
	connection.lastActivityJsonPayload = ""
}

func openDiscordPipe() (windows.Handle, error) {
	var lastError error
	for pipeIndex := 0; pipeIndex < pipeCandidateCount; pipeIndex++ {
		pathUTF16, conversionError := windows.UTF16FromString(fmt.Sprintf(discordPipePathTemplate, pipeIndex))
		if conversionError != nil {
			continue
		}

		pipeHandle, openError := windows.CreateFile(
			&pathUTF16[0],
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0,
			nil,
			windows.OPEN_EXISTING,
			0,
			0,
		)
		if openError == nil {
			return pipeHandle, nil
		}
		lastError = openError
	}
	return 0, fmt.Errorf("discord pipe not available: %w", lastError)
}

func performHandshake(pipeHandle windows.Handle, clientIdentifier string) error {
	payloadBytes, marshalError := json.Marshal(map[string]interface{}{
		"v":         1,
		"client_id": clientIdentifier,
	})
	if marshalError != nil {
		return marshalError
	}

	if writeError := writeFrame(pipeHandle, operationCodeHandshake, payloadBytes); writeError != nil {
		return writeError
	}

	responseLength, readHeaderError := readFrameLength(pipeHandle, handshakeResponseWait)
	if readHeaderError != nil {
		return readHeaderError
	}
	return readFrameBody(pipeHandle, responseLength, handshakeResponseWait)
}

func (connection *pipeConnection) ensureConnected(clientIdentifier string) error {
	if connection.isConnected && connection.currentClientIdentifier == clientIdentifier && connection.handle != 0 && connection.isAlive() {
		return nil
	}
	connection.resetLocked()

	pipeHandle, openError := openDiscordPipe()
	if openError != nil {
		return openError
	}

	if handshakeError := performHandshake(pipeHandle, clientIdentifier); handshakeError != nil {
		_ = windows.CloseHandle(pipeHandle)
		return handshakeError
	}

	connection.handle = pipeHandle
	connection.isConnected = true
	connection.currentClientIdentifier = clientIdentifier
	connection.lastActivityJsonPayload = ""
	return nil
}

func (connection *pipeConnection) sendActivity(clientIdentifier string, activity Activity) error {
	activityBytes, marshalActivityError := json.Marshal(activity)
	if marshalActivityError != nil {
		return marshalActivityError
	}
	activityJsonPayload := string(activityBytes)

	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	if connection.isConnected && connection.handle != 0 {
		if !connection.isAlive() {
			connection.resetLocked()
		} else if activityJsonPayload == connection.lastActivityJsonPayload {
			return nil
		}
	}

	if connectError := connection.ensureConnected(clientIdentifier); connectError != nil {
		return connectError
	}

	payloadBytes, marshalError := activityPayload(activity)
	if marshalError != nil {
		return marshalError
	}

	if writeError := writeFrame(connection.handle, operationCodeFrame, payloadBytes); writeError != nil {
		connection.resetLocked()
		return writeError
	}

	responseLength, readHeaderError := readFrameLength(connection.handle, activityResponseWait)
	if readHeaderError != nil {
		connection.resetLocked()
		return readHeaderError
	}
	_ = readFrameBody(connection.handle, responseLength, activityResponseWait)

	connection.lastActivityJsonPayload = activityJsonPayload
	return nil
}

func (connection *pipeConnection) clearActivity() {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	if !connection.isConnected || connection.handle == 0 {
		return
	}

	if payloadBytes, marshalError := activityPayload(nil); marshalError == nil {
		if writeError := writeFrame(connection.handle, operationCodeFrame, payloadBytes); writeError == nil {
			if responseLength, readHeaderError := readFrameLength(connection.handle, clearActivityWait); readHeaderError == nil && responseLength > 0 {
				_ = readFrameBody(connection.handle, responseLength, clearActivityWait)
			}
		}
	}

	connection.resetLocked()
}

func (connection *pipeConnection) close() {
	connection.clearActivity()
}
