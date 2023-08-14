package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/spf13/cobra"
)

type Setup struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Creates a Git Helper config file at ~/.git_helper/config.yml",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			setup().execute()
			return nil
		},
	}

	return cmd
}

func setup() *Setup {
	return &Setup{}
}

func (s *Setup) execute() {
	createConfig()
	setupPlugins()
}

func createConfig() {
	var create bool

	if configfile.ConfigFileExists() {
		create = commandline.AskYesNoQuestion("It looks like the " + configfile.ConfigFile() + " file already exists. Do you wish to replace it?")
	} else {
		create = true
	}

	if create {
		createOrUpdateConfig()
	}
}

func createOrUpdateConfig() {
	content := generateConfigFileContents()

	if !configfile.ConfigDirExists() {
		err := os.Mkdir(configfile.ConfigDir(), 0755)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	err := os.WriteFile(configfile.ConfigFile(), []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("\nDone setting up %s!\n\n", configfile.ConfigFile())
}

func generateConfigFileContents() string {
	var contents string

	github := commandline.AskYesNoQuestion("Do you wish to set up GitHub credentials?")

	if github {
		contents = contents + "github_username: " + commandline.AskOpenEndedQuestion("GitHub username?", false) + "\n"
		contents = contents + "github_token: " + commandline.AskOpenEndedQuestion("GitHub personal access token? (Navigatge to https://github.com/settings/tokens to create a new personal access token)", true) + "\n"
	}

	gitlab := commandline.AskYesNoQuestion("Do you wish to set up GitLab credentials?")

	if gitlab {
		contents = contents + "gitlab_username: " + commandline.AskOpenEndedQuestion("GitLab username?", false) + "\n"
		contents = contents + "gitlab_token: " + commandline.AskOpenEndedQuestion("GitLab personal access token? (Navigatge to https://gitlab.com/-/profile/personal_access_tokens to create a new personal access token)", true) + "\n"
	}

	contents = strings.TrimSpace(contents) + "\n"

	return contents
}

func setupPlugins() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up the Git Helper plugins?")

	if setup {
		createOrUpdatePlugins()
	}
}

func createOrUpdatePlugins() {
	pluginsDir := configfile.ConfigDir() + "/plugins"
	pluginsURL := "https://api.github.com/repos/emmahsax/git_helper/contents/plugins" // TODO: change this git repo when ready

	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		log.Fatal(err)
		return
	}

	resp, err := http.Get(pluginsURL)
	if err != nil {
		log.Fatal("Error:", err)
		return
	}
	defer resp.Body.Close()

	var allPlugins []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allPlugins); err != nil {
		log.Fatal(err)
		return
	}

	for _, plugin := range allPlugins {
		pluginURL := plugin["download_url"].(string)
		pluginName := plugin["name"].(string)

		resp, err := http.Get(pluginURL)
		if err != nil {
			log.Fatal(err)
			continue
		}
		defer resp.Body.Close()

		pluginPath := filepath.Join(pluginsDir, pluginName)
		file, err := os.Create(pluginPath)
		if err != nil {
			log.Fatal(err)
			continue
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Fatal(err)
			continue
		}
	}

	fmt.Printf("\nDone setting up plugins at %s!\n", pluginsDir)
	fmt.Println("\nNow add this line to your ~/.zshrc:\n  export PATH=\"$HOME/.git_helper/plugins:$PATH\"")
}
