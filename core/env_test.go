package core

import "testing"

func TestUserHomeDir_ReturnsNonEmptyString(t *testing.T) {
	// The os.UserHomeDir() error branch (fallback to "/root") is
	// impractical to force in-process, so only the happy path is
	// asserted here by design.
	got := UserHomeDir()
	if got == "" {
		t.Fatal("UserHomeDir() = \"\", want non-empty string")
	}
}
