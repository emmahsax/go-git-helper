package commandline

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

func AskMultipleChoice(question string, choices []string) string {
	selectedOption, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText(question).
		WithOnInterruptFunc(func() {
			os.Exit(1)
		}).
		WithOptions(choices).
		Show()

	return selectedOption
}

var AskOpenEndedQuestion = func(question string, secret bool) string {
	var result string
	for {
		if secret {
			result, _ = pterm.DefaultInteractiveTextInput.
				WithMultiLine(false).
				WithDefaultText(question).
				WithMask("*").
				WithOnInterruptFunc(func() {
					os.Exit(1)
				}).
				Show()
		} else {
			result, _ = pterm.DefaultInteractiveTextInput.
				WithMultiLine(false).
				WithDefaultText(question).
				WithOnInterruptFunc(func() {
					os.Exit(1)
				}).
				Show()
		}

		if result != "" {
			break
		}

		fmt.Println("--- This question is required ---")
	}

	return result
}

var AskYesNoQuestion = func(question string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText(question).
		WithDefaultValue(true).
		WithOnInterruptFunc(func() {
			os.Exit(1)
		}).
		Show()

	return result
}
