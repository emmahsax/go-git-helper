package newBranch

import (
	"io"
	"os"
	"testing"
)

func TestIsValidBranch(t *testing.T) {
	output := isValidBranch("hello-world")

	if output == false {
		t.Fatalf(`Branch %s should be valid`, "hello-world")
	}

	output = isValidBranch("hello_world")

	if output == false {
		t.Fatalf(`Branch %s should be valid`, "hello_world")
	}

	output = isValidBranch("hello world")

	if output {
		t.Fatalf(`Branch %s should be invalid`, "hello world")
	}

	output = isValidBranch("hello_world!")

	if output {
		t.Fatalf(`Branch %s should be invalid`, "hello_world!")
	}

	output = isValidBranch("hello_world?")

	if output {
		t.Fatalf(`Branch %s should be invalid`, "hello_world?")
	}

	output = isValidBranch("#helloWorld")

	if output {
		t.Fatalf(`Branch %s should be invalid`, "#helloWorld")
	}

	output = isValidBranch("hello-world*")

	if output {
		t.Fatalf(`Branch %s should be invalid`, "hello-world*")
	}
}

func TestGetValidBranch(t *testing.T) {
	testCases := []struct {
		expected string
		input    string
	}{
		{
			expected: "hello-world",
			input:    "hello-world\n",
		},
	}

	for _, tc := range testCases {
		t.Run("New branch name?", func(t *testing.T) {
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

			response := getValidBranch()

			if response != tc.expected {
				t.Errorf("Expected response %s, but got %s", tc.expected, response)
			}
		})
	}
}
