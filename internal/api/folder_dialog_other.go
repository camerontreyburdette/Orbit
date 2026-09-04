//go:build !windows

package api

func OpenFolderDialog(title string, parentWindow uintptr) (string, error) {
	return "", nil
}
