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

	pterm.Println()
	pterm.Info.Printfln("Selected option: %s", pterm.Green(selectedOption))
	pterm.Println()

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

	pterm.Println()

	if result == "" {
		fmt.Printf("--- This question is required ---\n\n")
		return AskOpenEndedQuestion(question, secret)
	}

	if secret == false {
		pterm.Info.Printfln("You answered: %s", result)
		pterm.Println()
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

	pterm.Println()
	pterm.Info.Printfln("You answered: %s", boolToText(result))
	pterm.Println()

	return result
}

func boolToText(b bool) string {
	if b {
		return pterm.Green("Yes")
	}
	return pterm.Red("No")
}
