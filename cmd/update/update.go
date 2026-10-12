package update

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// httpClient bounds how long a stalled connection can hang the update.
var httpClient = &http.Client{Timeout: 5 * time.Minute}

type Update struct {
	Debug       bool
	Executable  func() (string, error)
	Executor    executor.ExecutorInterface
	GOARCH      string
	GOOS        string
	InstallPath string
	LookPath    func(string) (string, error)
	Owner       string
	Repository  string
}

func NewCommand(packageOwner, packageRepository string) *cobra.Command {
	var (
		debug       bool
		installPath string
	)

	cmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Updates Git Helper with the newest version on GitHub",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := newUpdate(packageOwner, packageRepository, debug, executor.NewExecutor(debug))
			u.InstallPath = installPath
			u.execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")
	cmd.Flags().StringVar(&installPath, "install-path", "", "path of the git-helper binary to replace (defaults to the currently running binary)")

	return cmd
}

func newUpdate(owner, repository string, debug bool, executor executor.ExecutorInterface) *Update {
	return &Update{
		Debug:      debug,
		Executable: os.Executable,
		Executor:   executor,
		GOARCH:     runtime.GOARCH,
		GOOS:       runtime.GOOS,
		LookPath:   exec.LookPath,
		Owner:      owner,
		Repository: repository,
	}
}

func (u *Update) execute() {
	target, err := u.installPath()
	if err != nil {
		utils.HandleError(err, u.Debug, nil)
		return
	}

	u.removeStaleBinary(target)

	fmt.Printf("Installing latest git-helper version to %s\n", target)

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.Owner, u.Repository)
	body, err := u.fetchReleaseBody(releaseURL)
	if err != nil {
		utils.HandleError(err, u.Debug, nil)
		return
	}

	downloadURL, err := u.getDownloadURL(body)
	if err != nil {
		utils.HandleError(err, u.Debug, nil)
		return
	}

	downloaded, err := u.downloadBinary(downloadURL)
	if err != nil {
		utils.HandleError(err, u.Debug, nil)
		return
	}
	defer os.Remove(downloaded)

	if err = u.install(downloaded, target); err != nil {
		utils.HandleError(err, u.Debug, nil)
		return
	}

	u.outputNewVersion(target)
}

// assetName is the release asset for a platform, e.g. git-helper_linux_amd64 or
// git-helper_windows_arm64.exe.
func assetName(goos, goarch string) string {
	name := "git-helper_" + goos + "_" + goarch
	if goos == "windows" {
		name += ".exe"
	}

	return name
}

// installPath is the binary to replace: --install-path if given, otherwise the
// binary that is currently running. When running through `go run`, the running
// binary is a throwaway build, so the installed git-helper on PATH is used.
func (u *Update) installPath() (string, error) {
	if u.InstallPath != "" {
		return filepath.Abs(u.InstallPath)
	}

	exe, err := u.Executable()
	if err != nil {
		return "", err
	}

	if isGoRunBinary(exe) {
		onPath, err := u.LookPath("git-helper")
		if err != nil {
			return "", fmt.Errorf("running through `go run` and no installed git-helper was found on PATH; pass --install-path to say where to install it")
		}

		exe = onPath
	}

	return filepath.EvalSymlinks(exe)
}

// isGoRunBinary reports whether path is a throwaway binary built by `go run`.
// Depending on the Go version it lives either in a go-build<digits> directory
// under the temp dir, or in the build cache (a directory named go-build, e.g.
// ~/Library/Caches/go-build/96/<hash>-d/main).
func isGoRunBinary(path string) bool {
	for _, part := range strings.Split(strings.ReplaceAll(path, `\`, "/"), "/") {
		if goBuildDir.MatchString(part) {
			return true
		}
	}

	return false
}

var goBuildDir = regexp.MustCompile(`^go-build\d*$`)

// removeStaleBinary deletes the renamed old binary left behind by a previous
// Windows update (a running .exe can be renamed but not overwritten).
func (u *Update) removeStaleBinary(target string) {
	if u.GOOS != "windows" {
		return
	}

	_ = os.Remove(target + ".old")
}

func (u *Update) fetchReleaseBody(releaseURL string) ([]byte, error) {
	resp, err := httpClient.Get(releaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: unexpected status %s", releaseURL, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (u *Update) getDownloadURL(body []byte) (string, error) {
	name := assetName(u.GOOS, u.GOARCH)
	downloadURL := gjson.GetBytes(body, "assets.#(name==\""+name+"\").browser_download_url").String()
	if downloadURL == "" {
		return "", fmt.Errorf("the latest release has no %s asset (unsupported platform %s/%s?)", name, u.GOOS, u.GOARCH)
	}

	return downloadURL, nil
}

// downloadBinary saves the release asset to a temporary file and returns its path.
func (u *Update) downloadBinary(downloadURL string) (string, error) {
	resp, err := httpClient.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: unexpected status %s", downloadURL, resp.Status)
	}

	out, err := os.CreateTemp("", "git-helper-download-*")
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		os.Remove(out.Name())
		return "", err
	}

	return out.Name(), nil
}

// install replaces target with the downloaded binary. The new binary is first
// staged next to the target so the final step is an atomic rename.
func (u *Update) install(downloaded, target string) error {
	staged, err := stageBinary(downloaded, filepath.Dir(target))
	if err != nil {
		if u.GOOS != "windows" && errors.Is(err, fs.ErrPermission) {
			return u.sudoInstall(downloaded, target)
		}

		return err
	}
	defer os.Remove(staged)

	if u.GOOS == "windows" {
		return replaceRunningBinary(staged, target)
	}

	return os.Rename(staged, target)
}

func stageBinary(src, dir string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.CreateTemp(dir, ".git-helper-update-*")
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(out.Name())
		return "", err
	}

	if err = out.Close(); err != nil {
		os.Remove(out.Name())
		return "", err
	}

	if err = os.Chmod(out.Name(), 0755); err != nil {
		os.Remove(out.Name())
		return "", err
	}

	return out.Name(), nil
}

// replaceRunningBinary swaps in the new binary on Windows, where the running
// executable cannot be overwritten but can be renamed.
func replaceRunningBinary(staged, target string) error {
	old := target + ".old"
	_ = os.Remove(old)

	if err := os.Rename(target, old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := os.Rename(staged, target); err != nil {
		_ = os.Rename(old, target)
		return err
	}

	return nil
}

func (u *Update) sudoInstall(downloaded, target string) error {
	fmt.Printf("%s is not writable by the current user; using sudo to install it\n", filepath.Dir(target))

	args := []string{"install", "-m", "0755"}

	// Without this, sudo leaves the binary owned by root. Keep the current user
	// and their primary group as the owner, like any binary they installed themselves.
	if owner, group, err := currentOwner(); err == nil {
		args = append(args, "-o", owner, "-g", group)
	}

	args = append(args, downloaded, target)

	_, err := u.Executor.Exec("waitAndStdout", "sudo", args...)
	if err != nil {
		return fmt.Errorf("installing %s with sudo: %w (run git-helper update --install-path with a writable path to avoid sudo)", target, err)
	}

	return nil
}

// currentOwner is the current user and their primary group, by name where the
// group can be looked up and by numeric id otherwise (install accepts both).
func currentOwner() (string, string, error) {
	current, err := user.Current()
	if err != nil {
		return "", "", err
	}

	group := current.Gid
	if g, err := user.LookupGroupId(current.Gid); err == nil {
		group = g.Name
	}

	return current.Username, group, nil
}

func (u *Update) outputNewVersion(target string) {
	output, err := u.Executor.Exec("actionAndOutput", target, "version")
	if err != nil {
		utils.HandleError(err, u.Debug, nil)
		return
	}
	fmt.Printf("Installed %s", string(output))
}
