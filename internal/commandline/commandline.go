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

// AskMultiSelect asks question with a checkbox-style list: zero, one, or
// several of choices can be picked (space to toggle, enter to confirm).
var AskMultiSelect = func(question string, choices []string) []string {
	selectedOptions, _ := pterm.DefaultInteractiveMultiselect.
		WithDefaultText(question).
		WithOnInterruptFunc(func() {
			os.Exit(1)
		}).
		WithOptions(choices).
		Show()

	return selectedOptions
}

// AskOpenEndedQuestion asks question, pre-filling defaultVal if it's
// non-empty (accepted as-is on Enter). If required is true, a blank answer
// (no defaultVal, nothing typed) re-asks the question instead of returning;
// if false, a blank answer is returned immediately as a valid "no answer"
// (e.g. "leave this blank to skip").
var AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
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

		if result != "" || !required {
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
