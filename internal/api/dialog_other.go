//go:build !windows

package api

func OpenFileDialog(allowMultipleSelection bool, parentWindow uintptr) ([]string, error) {
	return []string{}, nil
}

func SaveFileDialog(defaultFilename string, parentWindow uintptr) (string, error) {
	return "", nil
}
