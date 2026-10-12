package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/emmahsax/go-git-helper/plugins"
	"github.com/spf13/cobra"
)

type Setup struct {
	Debug    bool
	Executor executor.ExecutorInterface
	Config   configfile.ConfigFileInterface
	GOOS     string
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Creates a Git Helper config file at ~/.git-helper/config.yml",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newSetup(debug, executor.NewExecutor(debug), configfile.NewConfigFile(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newSetup(debug bool, executor executor.ExecutorInterface, config configfile.ConfigFileInterface) *Setup {
	return &Setup{
		Debug:    debug,
		Executor: executor,
		Config:   config,
		GOOS:     runtime.GOOS,
	}
}

func (s *Setup) execute() {
	s.setupConfig()
	s.setupPlugins()
	s.setupCompletion()
}

func (s *Setup) setupConfig() {
	var create bool

	if s.Config.ConfigFileExists() {
		create = commandline.AskYesNoQuestion("The " + s.Config.ConfigFile() + " file already exists. Do you wish to replace it?")
	} else {
		create = true
	}

	if create {
		s.createOrUpdateConfig()
	}
}

func (s *Setup) createOrUpdateConfig() {
	content := s.generateConfigFileContents()

	if !s.Config.ConfigDirExists() {
		err := os.MkdirAll(s.Config.ConfigDir(), 0700)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
	}

	// Also tighten a directory created by an older version with looser permissions.
	if err := os.Chmod(s.Config.ConfigDir(), 0700); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	// The config file holds access tokens, so keep it private to the user. The
	// explicit chmod also tightens a file left over from an older version.
	err := os.WriteFile(s.Config.ConfigFile(), []byte(content), 0600)
	if err == nil {
		err = os.Chmod(s.Config.ConfigFile(), 0600)
	}
	if err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	fmt.Printf("\nDone setting up %s!\n\n", s.Config.ConfigFile())
}

func (s *Setup) generateConfigFileContents() string {
	var contents string

	github := commandline.AskYesNoQuestion("Do you wish to set up GitHub credentials?")

	if github {
		contents = contents + "github_username: " + commandline.AskOpenEndedQuestion("GitHub username", "", false) + "\n"
		contents = contents + "github_token: " + commandline.AskOpenEndedQuestion("GitHub personal access token - navigate to https://github.com/settings/tokens to create a new personal access token", "", true) + "\n"
	}

	gitlab := commandline.AskYesNoQuestion("Do you wish to set up GitLab credentials?")

	if gitlab {
		contents = contents + "gitlab_username: " + commandline.AskOpenEndedQuestion("GitLab username", "", false) + "\n"
		contents = contents + "gitlab_token: " + commandline.AskOpenEndedQuestion("GitLab personal access token - navigate to https://gitlab.com/-/profile/personal_access_tokens to create a new personal access token", "", true) + "\n"
	}

	contents = strings.TrimSpace(contents) + "\n"

	return contents
}

func (s *Setup) setupPlugins() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up the Git Helper plugins?")

	if setup {
		s.createOrUpdatePlugins()
	}
}

func (s *Setup) createOrUpdatePlugins() {
	pluginsDir := filepath.Join(s.Config.ConfigDir(), "plugins")

	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	entries, err := plugins.FS.ReadDir(".")
	if err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	for _, entry := range entries {
		name := entry.Name()

		content, err := plugins.FS.ReadFile(name)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}

		if err := writeExecutable(filepath.Join(pluginsDir, name), lfLineEndings(content)); err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}

		// cmd.exe and PowerShell can't run the sh script, so also provide a shim
		// that they will find through PATHEXT.
		if s.GOOS == "windows" {
			shim := "@git-helper " + strings.TrimPrefix(name, "git-") + " %*\r\n"
			if err := writeExecutable(filepath.Join(pluginsDir, name+".cmd"), []byte(shim)); err != nil {
				utils.HandleError(err, s.Debug, nil)
				return
			}
		}
	}

	fmt.Printf("\nDone setting up plugins at %s\n", pluginsDir)
	fmt.Printf("\n%s\n\n", pluginPathInstructions(s.GOOS, pluginsDir))
}

// lfLineEndings strips carriage returns so the sh scripts still run when the
// binary was built from a checkout that converted them to CRLF (as Git for
// Windows can), where `#!/bin/sh\r` would fail to execute.
func lfLineEndings(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
}

func writeExecutable(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0755); err != nil {
		return err
	}

	return os.Chmod(path, 0755)
}

func pluginPathInstructions(goos, pluginsDir string) string {
	if goos == "windows" {
		return "Now add the plugins directory to your PATH. In PowerShell:\n" +
			"  $dir = \"" + pluginsDir + "\"\n" +
			"  $parts = @([Environment]::GetEnvironmentVariable(\"Path\", \"User\") -split \";\" | Where-Object { $_ })\n" +
			"  if ($parts -notcontains $dir) { [Environment]::SetEnvironmentVariable(\"Path\", (($parts + $dir) -join \";\"), \"User\") }\n" +
			"Then open a new terminal"
	}

	dir := shellPath(pluginsDir)

	return "Now add this line to your shell's startup file (e.g. ~/.zshrc or ~/.bashrc):\n  export PATH=\"" + dir + ":$PATH\"\nFor fish, run:\n  fish_add_path \"" + dir + "\""
}

// shellPath writes a path under the home directory as $HOME/..., so the line
// users copy into their shell startup files also works when those are shared
// between machines. Other paths are returned unchanged.
func shellPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	rel, err := filepath.Rel(home, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}

	return "$HOME/" + filepath.ToSlash(rel)
}

func (s *Setup) setupCompletion() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up Git Helper completion?")

	if setup {
		s.createOrUpdateCompletion()
	}
}

func (s *Setup) createOrUpdateCompletion() {
	shes := []string{"bash", "fish", "powershell", "zsh"}
	completionsDir := filepath.Join(s.Config.ConfigDir(), "completions")
	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	for _, sh := range shes {
		output, err := s.Executor.Exec("actionAndOutput", "git-helper", "completion", sh)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}

		if err := os.WriteFile(filepath.Join(completionsDir, "completion."+sh), output, 0644); err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
	}

	fmt.Printf("\nCompletions (for bash, fish, powershell, and zsh) generated in %s\n\n%s\n\n", completionsDir, completionInstructions(s.GOOS, completionsDir))
}

func completionInstructions(goos, completionsDir string) string {
	if goos == "windows" {
		return "To activate completion in PowerShell, add this line to your profile (run `notepad $PROFILE` to open it):\n  . \"" + filepath.Join(completionsDir, "completion.powershell") + "\""
	}

	file := func(sh string) string {
		return "\"" + shellPath(filepath.Join(completionsDir, "completion."+sh)) + "\""
	}

	return "To activate completion, add the line for your shell to its startup file:\n" +
		"  zsh  (~/.zshrc):                   [ -f " + file("zsh") + " ] && source " + file("zsh") + "\n" +
		"  bash (~/.bashrc):                  [ -f " + file("bash") + " ] && source " + file("bash") + "\n" +
		"  fish (~/.config/fish/config.fish): test -f " + file("fish") + "; and source " + file("fish")
}
