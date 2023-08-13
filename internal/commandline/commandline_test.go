package commandline

import (
	"io"
	"os"
	"testing"
)

func TestAskMultipleChoice(t *testing.T) {
	testCases := []struct {
		question string
		expected string
		input    string
		choices  []string
	}{
		{
			question: "What is your favorite color?",
			expected: "Blue",
			input:    "2\n",
			choices:  []string{"Red", "Blue", "Green"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.question, func(t *testing.T) {
			reader, writer, _ := os.Pipe()
			defer reader.Close()
			defer writer.Close()

			originalStdin := os.Stdin
			os.Stdin = reader
			defer func() {
				os.Stdin = originalStdin
			}()

			_, err := io.WriteString(writer, tc.input)
			if err != nil {
				t.Fatal(err)
			}

			response := AskMultipleChoice(tc.question, tc.choices)

			if response != tc.expected {
				t.Errorf("Expected response %s, but got %s", tc.expected, response)
			}
		})
	}
}

func TestAskOpenEndedQuestion(t *testing.T) {
	testCases := []struct {
		question string
		expected string
		input    string
		secret   bool
	}{
		{
			question: "What is your favorite animal?",
			expected: "Lion",
			input:    "Lion\n",
			secret:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.question, func(t *testing.T) {
			reader, writer, _ := os.Pipe()
			defer reader.Close()
			defer writer.Close()

			originalStdin := os.Stdin
			os.Stdin = reader
			defer func() {
				os.Stdin = originalStdin
			}()

			_, err := io.WriteString(writer, tc.input)
			if err != nil {
				t.Fatal(err)
			}

			response := AskOpenEndedQuestion(tc.question, tc.secret)

			if response != tc.expected {
				t.Errorf("Expected response %s, but got %s", tc.expected, response)
			}
		})
	}
}

func TestAskYesNoQuestion(t *testing.T) {
	testCases := []struct {
		question string
		expected bool
		input    string
	}{
		{
			question: "Do you like purple?",
			expected: true,
			input:    "yes\n",
		},
		{
			question: "Do you prefer green?",
			expected: false,
			input:    "no\n",
		},
		{
			question: "What about orange?",
			expected: true,
			input:    "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.question, func(t *testing.T) {
			reader, writer, _ := os.Pipe()
			defer reader.Close()
			defer writer.Close()

			originalStdin := os.Stdin
			os.Stdin = reader
			defer func() {
				os.Stdin = originalStdin
			}()

			_, err := io.WriteString(writer, tc.input)
			if err != nil {
				t.Fatal(err)
			}

			response := AskYesNoQuestion(tc.question)

			if response != tc.expected {
				t.Errorf("Expected response %v, but got %v", tc.expected, response)
			}
		})
	}
}
