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

func AskOpenEndedQuestion(question string, secret bool) string {
	var result string

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

	if result == "" {
		fmt.Println("--- This question is required ---")
		return AskOpenEndedQuestion(question, secret)
	}

	return result
}

func AskYesNoQuestion(question string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText(question).
		WithDefaultValue(true).
		WithOnInterruptFunc(func() {
			os.Exit(1)
		}).
		Show()

	return result
}

func boolToText(b bool) string {
	if b {
		return pterm.Green("Yes")
	}
	return pterm.Red("No")
}
