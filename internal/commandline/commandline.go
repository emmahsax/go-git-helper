package commandline

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

var AskMultipleChoice = func(question string, choices []string) string {
	selectedOption, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText(question).
		WithOnInterruptFunc(func() {
			os.Exit(1)
		}).
		WithOptions(choices).
		Show()

	return selectedOption
}

var AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string {
	var result string
	for {
		if secret {
			result, _ = pterm.DefaultInteractiveTextInput.
				WithDefaultText(question).
				WithMask("*").
				WithMultiLine(false).
				WithOnInterruptFunc(func() {
					os.Exit(1)
				}).
				Show()
		} else if defaultVal == "" {
			result, _ = pterm.DefaultInteractiveTextInput.
				WithDefaultText(question).
				WithMultiLine(false).
				WithOnInterruptFunc(func() {
					os.Exit(1)
				}).
				Show()
		} else {
			result, _ = pterm.DefaultInteractiveTextInput.
				WithDefaultText(question).
				WithDefaultValue(defaultVal).
				WithMultiLine(false).
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
