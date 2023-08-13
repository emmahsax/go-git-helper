package changeRemote

import (
	"fmt"
	"log"
	"net/url"
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
		Use:   "change-remote [oldOwner] [newOwner]",
		Short: "Change the git remote owners for multiple cloned git repositories",
		Args:  cobra.ExactArgs(2),
		DisableFlagParsing: true,
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
	}

	remotes := strings.Split(string(output), "\n")

	for _, remote := range remotes {
		if strings.Contains(remote, cr.OldOwner) {
			processRemote(remote, cr)
		} else {
			fmt.Printf("  Found remote is not pointing to %s.\n", cr.OldOwner)
		}
	}
	fmt.Println()
}

func processRemote(remote string, cr *ChangeRemote) {
	remoteName := strings.Split(remote, " ")[0]
	var newRemote string

	if strings.Contains(remote, "git@") {
		host, _, repo := remote_info(remote)
		newRemote = fmt.Sprintf("git@%s:%s/%s.git", host, cr.NewOwner, repo)
	} else if strings.Contains(remote, "https://") {
		host, _, repo := remote_info(remote)
		newRemote = fmt.Sprintf("https://%s/%s/%s.git", host, cr.NewOwner, repo)
	}

	fmt.Printf("  Changing the remote URL %s to be '%s'.\n", remote, newRemote)
	cmd := exec.Command("git", "remote", "set-url", remoteName, newRemote)
	_, err := cmd.Output()
	if err != nil {
		log.Fatal("Error:", err)
	}
}

func remote_info(remote string) (string, string, string) {
	if strings.Contains(remote, "git@") {
		hostAndOrgRepo := strings.SplitN(remote, ":", 2)
		if len(hostAndOrgRepo) != 2 {
			log.Fatal("Invalid remote URL format")
		}

		parts := strings.SplitN(hostAndOrgRepo[1], "/", 2)
		if len(parts) != 2 {
			log.Fatal("Invalid remote URL format")
		}

		return parts[0], strings.Split(parts[1], "/")[0], strings.Split(parts[1], "/")[1]
	} else if strings.Contains(remote, "https://") {
			parsedURL, err := url.Parse(remote)
		if err != nil {
			log.Fatal("Error parsing URL:", err)
		}

		path := filepath.Clean(parsedURL.Path)
		pathParts := strings.Split(path, "/")

		if len(pathParts) < 2 {
			log.Fatal("Invalid remote URL format")
		}

		return parsedURL.Host, pathParts[1], strings.TrimSuffix(pathParts[len(pathParts)-1], ".git")
	} else {
		log.Fatal("Invalid remote URL format")
		return "", "", ""
	}
}
