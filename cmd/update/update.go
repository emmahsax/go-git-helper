package update

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type Update struct{}

var (
	asset      = "git-helper_darwin_arm64"
	owner      = "emmahsax"
	repository = "go-git-helper"
	newPath    = "/usr/local/bin/go-git-helper" // TODO: switch this to git-helper, it's leftover from the migration
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Updates Git Helper with the newest version on GitHub",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newUpdate().execute()
			return nil
		},
	}

	return cmd
}

func newUpdate() *Update {
	return &Update{}
}

func (u *Update) execute() {
	downloadGitHelper()
	moveGitHelper()
	setPermissions()
	outputNewVersion()
}

func downloadGitHelper() {
	fmt.Println("Installing latest git-helper version")

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repository)
	resp, err := http.Get(releaseURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	var downloadURL string
	switch runtime.GOOS {
	case "darwin":
		downloadURL = gjson.Get(string(body), "assets.#(name==\""+asset+"\").browser_download_url").String()
	default:
		fmt.Println("Unsupported operating system:", runtime.GOOS)
		return
	}

	binaryName := strings.Split(downloadURL, "/")[len(strings.Split(downloadURL, "/"))-1]
	resp, err = http.Get(downloadURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(binaryName)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func moveGitHelper() {
	cmd := exec.Command("sudo", "mv", "./"+asset, newPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func setPermissions() {
	cmdChown := exec.Command("sudo", "chown", "root:wheel", newPath)
	output, err := cmdChown.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))

	cmdChmod := exec.Command("sudo", "chmod", "+x", newPath)
	output, err = cmdChmod.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func outputNewVersion() {
	cmd := exec.Command("go-git-helper", "version") // TODO: make this command git-helper
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("Installed %s", string(output))
}
