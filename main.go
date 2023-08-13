package main

import (
	"fmt"

	"github.com/emmahsax/go-git-helper/internal/commandline"
)

func main() {
	fmt.Println("hello world")

	// choices := []string{"Blue", "Yellow", "Orange", "Red"}
	// fmt.Println(commandline.AskMultipleChoice("What's your favorite color?", choices))

	fmt.Println(commandline.AskOpenEndedQuestion("What's your password?", true))

	// fmt.Println(commandline.AskYesNoQuestion("Is Go your new favorite language?"))
}
