package executor

import (
	"errors"
	"os"
	"os/exec"
)

type ExecutorInterface interface {
	Exec(execType string, command string, args ...string) ([]byte, error)
}

type Executor struct {
	Args    []string
	Command string
	Debug   bool
}

func NewExecutor(debug bool) *Executor {
	return &Executor{
		Debug: debug,
	}
}

func (e *Executor) Exec(execType string, command string, args ...string) ([]byte, error) {
	e.Command = command
	e.Args = args

	switch execType {
	case "actionAndOutput":
		o, e := actionAndOutput(command, args)
		return o, e
	case "waitAndStdout":
		return []byte{}, waitAndStdout(command, args)
	default:
		return []byte{}, errors.New("invalid exec type")
	}
}

func actionAndOutput(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func waitAndStdout(command string, args []string) error {
	origStdout := os.Stdout
	origStderr := os.Stderr

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	os.Stdout = origStdout
	os.Stderr = origStderr

	return nil
}
