package changeRemote

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/spf13/cobra"
)

type ChangeRemote struct {
	Debug    bool
	Executor executor.ExecutorInterface
	NewOwner string
	OldOwner string
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "change-remote [oldOwner] [newOwner]",
		Short:                 "Change the git remote owners for multiple cloned git repositories",
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newChangeRemote(args[0], args[1], debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newChangeRemote(oldOwner, newOwner string, debug bool) *ChangeRemote {
	return &ChangeRemote{
		Debug:    debug,
		Executor: executor.NewExecutor(debug),
		NewOwner: newOwner,
		OldOwner: oldOwner,
	}
}

func (cr *ChangeRemote) execute() {
	originalDir, _ := os.Getwd()
	nestedDirs, _ := os.ReadDir(originalDir)

	for _, entry := range nestedDirs {
		if entry.IsDir() && entry.Name() != "." && entry.Name() != ".." {
			cr.processDir(entry.Name(), originalDir)
		}
	}
}

func (cr *ChangeRemote) processDir(currentDir, originalDir string) {
	_ = os.Chdir(currentDir)
	defer os.Chdir(originalDir)

	current, _ := os.Getwd()
	gitDir := filepath.Join(current, ".git")

	if _, err := os.Stat(gitDir); err == nil {
		fullRemoteInfo := cr.processGitRepository()

		if len(fullRemoteInfo) > 0 {
			fmt.Println("Found git directory: " + currentDir + ".")
		}

		for repo, inner := range fullRemoteInfo {
			answer := commandline.AskYesNoQuestion(
				"Do you wish to proceed in updating the " + inner["plainRemote"] + " remote URL?",
			)

			if answer {
				fmt.Println("doing something different")
				cr.processRemote(inner["plainRemote"], inner["host"], repo, inner["remoteName"])
			}
		}
	}
}

func (cr *ChangeRemote) processGitRepository() map[string]map[string]string {
	fullRemoteInfo := make(map[string]map[string]string)

	output, err := cr.Executor.Exec("git", "remote", "-v")
	if err != nil {
		log.Fatal(err)
		return fullRemoteInfo
	}

	remotes := strings.Split(string(output), "\n")

	for _, remote := range remotes {
		if remote == "" {
			continue
		}

		plainRemote := strings.Fields(remote)[1]
		remoteName := strings.Fields(remote)[0]
		host, owner, repo := cr.remoteInfo(plainRemote)

		if owner == cr.OldOwner {
			inner := make(map[string]string)
			inner["plainRemote"] = plainRemote
			inner["remoteName"] = remoteName
			inner["host"] = host
			fullRemoteInfo[repo] = inner
		}
	}
	return fullRemoteInfo
}

func (cr *ChangeRemote) processRemote(remote, host, repo, remoteName string) {
	var newRemote string

	if strings.Contains(remote, "git@") {
		newRemote = fmt.Sprintf("%s:%s/%s", host, cr.NewOwner, repo)
	} else if strings.Contains(remote, "https://") {
		newRemote = fmt.Sprintf("%s/%s/%s", host, cr.NewOwner, repo)
	}

	fmt.Printf("  Changing the remote URL '%s' to be '%s'.\n", remote, newRemote)

	output, err := cr.Executor.Exec("git", "remote", "set-url", remoteName, newRemote)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(string(output))
}

func (cr *ChangeRemote) remoteInfo(remote string) (string, string, string) {
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
