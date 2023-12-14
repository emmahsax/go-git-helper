package newBranch

import (
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
