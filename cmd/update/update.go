package update

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type Update struct {
	Debug    bool
	Executor executor.ExecutorInterface
}

var (
	asset      = "git-helper_" + runtime.GOOS + "_" + runtime.GOARCH
	owner      = "emmahsax"
	repository = "go-git-helper"
	newPath    = "/usr/local/bin/git-helper" // This is for linux and mac based systems only
)

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Updates Git Helper with the newest version on GitHub",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newUpdate(debug, executor.NewExecutor(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newUpdate(debug bool, executor executor.ExecutorInterface) *Update {
	return &Update{
		Debug:    debug,
		Executor: executor,
	}
}

func (u *Update) execute() {
	u.downloadGitHelper()
	u.moveGitHelper()
	u.setPermissions()
	u.outputNewVersion()
}

func (u *Update) downloadGitHelper() {
	fmt.Println("Installing latest git-helper version")

	releaseURL := u.buildReleaseURL()
	body := u.fetchReleaseBody(releaseURL)
	downloadURL := u.getDownloadURL(body)
	binaryName := u.getBinaryName(downloadURL)
	u.downloadAndSaveBinary(downloadURL, binaryName)
}

func (u *Update) buildReleaseURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repository)
}

func (u *Update) fetchReleaseBody(releaseURL string) []byte {
	resp, err := http.Get(releaseURL)
	if err != nil {
		u.handleError(err)
		return []byte{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		u.handleError(err)
		return []byte{}
	}

	return body
}

func (u *Update) getDownloadURL(body []byte) string {
	var downloadURL string
	switch runtime.GOOS {
	case "darwin":
		downloadURL = gjson.Get(string(body), "assets.#(name==\""+asset+"\").browser_download_url").String()
	case "linux":
		downloadURL = gjson.Get(string(body), "assets.#(name==\""+asset+"\").browser_download_url").String()
	default:
		fmt.Println("Unsupported operating system:", runtime.GOOS)
	}

	return downloadURL
}

func (u *Update) getBinaryName(downloadURL string) string {
	return strings.Split(downloadURL, "/")[len(strings.Split(downloadURL, "/"))-1]
}

func (u *Update) downloadAndSaveBinary(downloadURL, binaryName string) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		u.handleError(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(binaryName)
	if err != nil {
		u.handleError(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		u.handleError(err)
	}
}

func (u *Update) moveGitHelper() {
	output, err := u.Executor.Exec("actionAndOutput", "sudo", "mv", "./"+asset, newPath)
	if err != nil {
		if u.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func (u *Update) setPermissions() {
	currentUser, err := user.Current()
	if err != nil {
		u.handleError(err)
		return
	}

	output, err := u.Executor.Exec("actionAndOutput", "sudo", "chown", currentUser.Username+":staff", newPath)
	if err != nil {
		u.handleError(err)
		return
	}

	fmt.Printf("%s", string(output))

	output, err = u.Executor.Exec("actionAndOutput", "sudo", "chmod", "+x", newPath)
	if err != nil {
		u.handleError(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func (u *Update) outputNewVersion() {
	output, err := u.Executor.Exec("actionAndOutput", "git-helper", "version")
	if err != nil {
		u.handleError(err)
		return
	}
	fmt.Printf("Installed %s", string(output))
}

func (u *Update) handleError(err error) {
	if u.Debug {
		debug.PrintStack()
	}
	log.Fatal(err)
}
