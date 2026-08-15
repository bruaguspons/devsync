package core

import "os"

// UserHomeDir returns the current user's home directory, falling back
// to "/root" if it cannot be determined — the shared fallback used by
// every on-disk default-path helper in this project.
func UserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/root"
	}
	return home
}
