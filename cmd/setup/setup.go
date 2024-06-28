package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/spf13/cobra"
)

type Setup struct {
	Debug      bool
	Executor   executor.ExecutorInterface
	Config     configfile.ConfigFileInterface
	Owner      string
	Repository string
}

func NewCommand(packageOwner, packageRepository string) *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Creates a Git Helper config file at ~/.git-helper/config.yml",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newSetup(packageOwner, packageRepository, debug, executor.NewExecutor(debug), configfile.NewConfigFile(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newSetup(owner, repository string, debug bool, executor executor.ExecutorInterface, config configfile.ConfigFileInterface) *Setup {
	return &Setup{
		Debug:      debug,
		Executor:   executor,
		Config:     config,
		Owner:      owner,
		Repository: repository,
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
		err := os.Mkdir(s.Config.ConfigDir(), 0755)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
	}

	err := os.WriteFile(s.Config.ConfigFile(), []byte(content), 0644)
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
		s.createOrUpdatePlugins(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/plugins", s.Owner, s.Repository))
	}
}

func (s *Setup) createOrUpdatePlugins(pluginsURL string) {
	pluginsDir := s.Config.ConfigDir() + "/plugins"

	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	resp, err := http.Get(pluginsURL)
	if err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}
	defer resp.Body.Close()

	var allPlugins []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allPlugins); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	for _, plugin := range allPlugins {
		pluginURL := plugin["download_url"].(string)
		pluginName := plugin["name"].(string)

		resp, err := http.Get(pluginURL)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			continue
		}
		defer resp.Body.Close()

		pluginPath := filepath.Join(pluginsDir, pluginName)
		file, err := os.Create(pluginPath)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			continue
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			continue
		}

		err = os.Chmod(pluginPath, 0755)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	fmt.Printf("\nDone setting up plugins at %s!\n", pluginsDir)
	fmt.Printf("\nNow add this line to your Unix shell file (e.g. ~/.zshrc):\n  export PATH=\"$HOME/.git-helper/plugins:$PATH\"\n\n")
}

func (s *Setup) setupCompletion() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up Git Helper completion?")

	if setup {
		s.createOrUpdateCompletion()
	}
}

func (s *Setup) createOrUpdateCompletion() {
	shes := []string{"bash", "fish", "powershell", "zsh"}
	completionsDir := s.Config.ConfigDir() + "/completions"
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

		filename := completionsDir + "/completion." + sh

		file, err := os.Create(filename)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
		defer file.Close()

		_, err = file.WriteString(string(output))
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
	}

	fmt.Println("\nCompletions (for bash, fish, powershell, and zsh) generated in " + completionsDir + ". Please activate the proper completion for your Unix shell. E.g. add the following to your ~/.zshrc file:\n  [ -f ~/.git-helper/completions/completion.zsh ] && source ~/.git-helper/completions/completion.zsh\n")
}
