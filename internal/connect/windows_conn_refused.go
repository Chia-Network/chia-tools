//go:build windows

package connect

import (
	"errors"

	"golang.org/x/sys/windows"
)

// IsWindowsConnectionRefused checks if the error is a Windows-specific connection refused error.
func IsWindowsConnectionRefused(err error) bool {
	return errors.Is(err, windows.WSAECONNREFUSED)
}
