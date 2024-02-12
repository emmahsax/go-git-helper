package setup

import (
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
)

type MockExecutor struct {
	Args    []string
	Command string
	Debug   bool
	Output  []byte
}

func (me *MockExecutor) Exec(execType string, command string, args ...string) ([]byte, error) {
	me.Command = command
	me.Args = args
	return me.Output, nil
}

func Test_generateConfigFileContents(t *testing.T) {
	// Mock the AskYesNoQuestion function to always return true
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	// Mock the AskOpenEndedQuestion function to always return "hello world"
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
		return "hello_world"
	}

	executor := &MockExecutor{Debug: true}
	s := newSetup(true, executor)
	contents := s.generateConfigFileContents()

	expectedContents := "github_username: hello_world\ngithub_token: hello_world\ngitlab_username: hello_world\ngitlab_token: hello_world\n"
	if contents != expectedContents {
		t.Errorf("Expected '%s', got '%s'", expectedContents, contents)
	}
}
