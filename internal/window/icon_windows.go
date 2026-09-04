//go:build windows

package window

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

const (
	iconResourceVersion = 0x00030000
	iconTypeIdentifier  = 1
	fullSizeIconWidth   = 256
)

type iconDirectoryHeaderStructure struct {
	Reserved uint16
	Type     uint16
	Count    uint16
}

type iconDirectoryEntryStructure struct {
	Width           uint8
	Height          uint8
	ColorCount      uint8
	Reserved        uint8
	Planes          uint16
	BitCount        uint16
	BytesInResource uint32
	ImageOffset     uint32
}

func readIconDirectoryEntries(iconData []byte) ([]iconDirectoryEntryStructure, bool) {
	if len(iconData) < 6 {
		return nil, false
	}

	var header iconDirectoryHeaderStructure
	reader := bytes.NewReader(iconData)
	if readError := binary.Read(reader, binary.LittleEndian, &header); readError != nil || header.Type != iconTypeIdentifier || header.Count == 0 {
		return nil, false
	}

	entries := make([]iconDirectoryEntryStructure, header.Count)
	for entryIndex := range entries {
		if readError := binary.Read(reader, binary.LittleEndian, &entries[entryIndex]); readError != nil {
			return nil, false
		}
	}
	return entries, true
}

func absoluteDifference(first int, second int) int {
	if first > second {
		return first - second
	}
	return second - first
}

func selectClosestIconEntry(entries []iconDirectoryEntryStructure, desiredWidth int) iconDirectoryEntryStructure {
	bestIndex := 0
	bestDifference := 99999
	for entryIndex, entry := range entries {
		entryWidth := int(entry.Width)
		if entryWidth == 0 {
			entryWidth = fullSizeIconWidth
		}
		if difference := absoluteDifference(entryWidth, desiredWidth); difference < bestDifference {
			bestDifference = difference
			bestIndex = entryIndex
		}
	}
	return entries[bestIndex]
}

func createIconFromIconData(iconData []byte, desiredWidth int, desiredHeight int) uintptr {
	entries, isValid := readIconDirectoryEntries(iconData)
	if !isValid {
		return 0
	}

	bestEntry := selectClosestIconEntry(entries, desiredWidth)
	startOffset := int(bestEntry.ImageOffset)
	endOffset := startOffset + int(bestEntry.BytesInResource)
	if startOffset < 0 || endOffset > len(iconData) || startOffset >= endOffset {
		return 0
	}

	imageData := iconData[startOffset:endOffset]
	iconHandle, _, _ := createIconFromResourceExtended.Call(
		uintptr(unsafe.Pointer(&imageData[0])),
		uintptr(len(imageData)),
		iconTypeIdentifier,
		iconResourceVersion,
		uintptr(desiredWidth),
		uintptr(desiredHeight),
		0,
	)
	return iconHandle
}
