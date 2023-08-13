package changeRemote

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/spf13/cobra"
)

type ChangeRemote struct {
	OldOwner string
	NewOwner string
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "change-remote [oldOwner] [newOwner]",
		Short:                 "Change the git remote owners for multiple cloned git repositories",
		Args:                  cobra.ExactArgs(2),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newChangeRemote(args[0], args[1]).execute()
			return nil
		},
	}

	return cmd
}

func newChangeRemote(oldOwner, newOwner string) *ChangeRemote {
	return &ChangeRemote{
		OldOwner: oldOwner,
		NewOwner: newOwner,
	}
}

func (cr *ChangeRemote) execute() {
	originalDir, _ := os.Getwd()
	nestedDirs, _ := os.ReadDir(originalDir)

	for _, entry := range nestedDirs {
		if entry.IsDir() && entry.Name() != "." && entry.Name() != ".." {
			processDir(entry.Name(), originalDir, cr)
		}
	}
}

func processDir(currentDir, originalDir string, cr *ChangeRemote) {
	_ = os.Chdir(currentDir)
	defer os.Chdir(originalDir)

	current, _ := os.Getwd()
	gitDir := filepath.Join(current, ".git")

	if _, err := os.Stat(gitDir); err == nil {
		answer := commandline.AskYesNoQuestion(
			"Found git directory: " + currentDir + ". Do you wish to proceed in updating " + currentDir + "'s remote URLs?",
		)

		if answer {
			processGitRepository(cr)
		}
	}
}

func processGitRepository(cr *ChangeRemote) {
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error:", err)
		return
	}

	remotes := strings.Split(string(output), "\n")

	for _, remote := range remotes {
		if remote == "" {
			continue
		}

		plainRemote := strings.Fields(remote)[1]
		remoteName := strings.Fields(remote)[0]
		host, owner, repo := remoteInfo(plainRemote)

		if owner == cr.OldOwner {
			processRemote(plainRemote, host, repo, remoteName, cr)
		} else {
			fmt.Printf("  Found remote (%s) is not pointing to %s.\n", plainRemote, cr.OldOwner)
		}
	}
	fmt.Println()
}

func processRemote(remote, host, repo, remoteName string, cr *ChangeRemote) {
	var newRemote string

	if strings.Contains(remote, "git@") {
		newRemote = fmt.Sprintf("%s:%s/%s", host, cr.NewOwner, repo)
	} else if strings.Contains(remote, "https://") {
		newRemote = fmt.Sprintf("%s/%s/%s", host, cr.NewOwner, repo)
	}

	fmt.Printf("  Changing the remote URL '%s' to be '%s'.\n", remote, newRemote)
	cmd := exec.Command("git", "remote", "set-url", remoteName, newRemote)
	_, err := cmd.Output()
	if err != nil {
		log.Fatal("Error:", err)
		return
	}
}

func remoteInfo(remote string) (string, string, string) {
	if strings.Contains(remote, "git@") {
		remoteSplit := strings.SplitN(remote, ":", 2)
		if len(remoteSplit) != 2 {
			log.Fatal("Invalid remote URL format")
			return "", "", ""
		}

		parts := strings.SplitN(remoteSplit[1], "/", 2)
		if len(parts) != 2 {
			log.Fatal("Invalid remote URL format")
			return "", "", ""
		}

		return remoteSplit[0], parts[0], parts[1]
	} else if strings.Contains(remote, "https://") {
		remoteSplit := strings.SplitN(remote, "/", -1)
		host := remoteSplit[0] + "//" + remoteSplit[2]

		return host, remoteSplit[3], remoteSplit[4]
	} else {
		log.Fatal("Invalid remote URL format")
		return "", "", ""
	}
}
