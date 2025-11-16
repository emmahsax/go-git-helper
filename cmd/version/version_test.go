package version

import (
	"testing"
)

func Test_NewCommand(t *testing.T) {
	cmd := NewCommand("1.2.3")

	if cmd.Use != "version" {
		t.Errorf("Expected command Use to be 'version', got '%s'", cmd.Use)
	}

	if cmd.Short != "Print the version number" {
		t.Errorf("Expected command Short to be 'Print the version number', got '%s'", cmd.Short)
	}
}

func Test_NewCommand_ExecuteWithVersion(t *testing.T) {
	version := "1.2.3"
	cmd := NewCommand(version)

	// Set args to empty to avoid default os.Args
	cmd.SetArgs([]string{})

	// Execute command - output goes to stdout, just check for no error
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// The actual output test is hard to capture due to fmt.Printf
	// We're verifying the command executes without error
}

func Test_NewCommand_WithDifferentVersion(t *testing.T) {
	version := "2.0.0-beta"
	cmd := NewCommand(version)

	// Set args to empty to avoid default os.Args
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// The actual output test is hard to capture due to fmt.Printf
	// We're verifying the command executes without error
}

func Test_NewCommand_NoArgs(t *testing.T) {
	cmd := NewCommand("1.0.0")

	// Set args to empty slice
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error with no args, got %v", err)
	}
}

func Test_NewCommand_WithArgs_ShouldFail(t *testing.T) {
	cmd := NewCommand("1.0.0")

	// Set args that should be rejected
	cmd.SetArgs([]string{"extra-arg"})

	err := cmd.Execute()
	if err == nil {
		t.Errorf("Expected error with extra args, got nil")
	}
}
