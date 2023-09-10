package changeRemote

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	assert.NoError(t, err)

	cr := newChangeRemoteClient("oldOwner", "newOwner", true)
	cr.execute()
}

func TestRemoteInfo(t *testing.T) {
	remote := "git@github.com:oldOwner/repo.git"
	cr := newChangeRemoteClient("oldOwner", "newOwner", false)
	cr.execute()
	host, owner, repo := cr.remoteInfo(remote)

	if host != "git@github.com" {
		t.Fatalf(`Host should match %s, was %s`, "git@github.com", host)
	}

	if owner != "oldOwner" {
		t.Fatalf(`Owner should match %s, was %s`, "oldOwner", host)
	}

	if repo != "repo.git" {
		t.Fatalf(`Repo should match %s, was %s`, "repo.git", host)
	}

	remote = "https://gitlab.com/oldOwner/repo"
	cr = newChangeRemoteClient("oldOwner", "newOwner", false)
	cr.execute()
	host, owner, repo = cr.remoteInfo(remote)

	if host != "https://gitlab.com" {
		t.Fatalf(`Host should match %s, was %s`, "https://gitlab.com", host)
	}

	if owner != "oldOwner" {
		t.Fatalf(`Owner should match %s, was %s`, "oldOwner", host)
	}

	if repo != "repo" {
		t.Fatalf(`Repo should match %s, was %s`, "repo", host)
	}
}
