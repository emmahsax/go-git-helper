package commandline

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func AskMultipleChoice(question string, choices []string) string {
	fmt.Println(question)

	for i, choice := range choices {
		fmt.Printf("%d. %s\n", i+1, choice)
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choiceNum, err := strconv.Atoi(input)
	if err != nil || choiceNum < 1 || choiceNum > len(choices) {
		fmt.Println("--- This question is required ---")
		AskMultipleChoice(question, choices)
	}

	return choices[choiceNum-1]
}

func AskOpenEndedQuestion(question string, secret bool) string {
	fmt.Println(question)
	var response string

	if secret {
		bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
		response = string(bytePassword)
		fmt.Println()
	} else {
		reader := bufio.NewReader(os.Stdin)
		response, _ = reader.ReadString('\n')
		response = strings.TrimSpace(response)
	}

	if response == "" {
		fmt.Println("--- This question is required ---")
		AskOpenEndedQuestion(question, secret)
	}

	return response
}

func AskYesNoQuestion(question string) bool {
	fmt.Printf("%s (Y/n): ", question)

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "yes" || response == "y" || response == "" {
		return true
	} else if response == "no" || response == "n" {
		return false
	} else {
		fmt.Println("--- This question is required ---")
		return AskYesNoQuestion(question)
	}
}
