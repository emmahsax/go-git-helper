package executor

import (
	"log"
	"os"
	"os/exec"
	"runtime/debug"
)

type ExecutorInterface interface {
	Exec(command string, args ...string) ([]byte, error)
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

func (e *Executor) Exec(command string, args ...string) ([]byte, error) {
	e.Command = command
	e.Args = args
	origStdout := os.Stdout
	origStdin := os.Stdin

	cmd := exec.Command(command, args...)
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	if e.Debug {
	// 		debug.PrintStack()
	// 	}
	// 	log.Fatal(err)
	// 	return nil, err
	// }

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		if e.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		if e.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return nil, err
	}

	os.Stdout = origStdout
	os.Stdin = origStdin
	return []byte{}, nil
}
