//go:build !windows

package connect

// IsWindowsConnectionRefused is a no-op on non-Windows systems.
func IsWindowsConnectionRefused(err error) bool {
	return false
}
