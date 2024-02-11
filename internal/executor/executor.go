package executor

import (
	"log"
	"os/exec"
	"runtime/debug"
)

type ExecutorInterface interface {
    Exec(command string, args ...string) ([]byte, error)
}

type Executor struct {
    Args []string
    Command string
    Debug bool
}

func NewExecutor(debug bool) *Executor {
	return &Executor{
		Debug:    debug,
	}
}

func (e *Executor) Exec(command string, args ...string) ([]byte, error) {
    e.Command = command
    e.Args = args

    cmd := exec.Command(command, args...)
    output, err := cmd.Output()
    if err != nil {
        if e.Debug {
            debug.PrintStack()
        }
        log.Fatal(err)
        return nil, err
    }
    return output, nil
}
