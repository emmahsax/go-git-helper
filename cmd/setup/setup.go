package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/spf13/cobra"
)

type Setup struct {
	Debug bool
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Creates a Git Helper config file at ~/.git_helper/config.yml",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newSetupClient(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newSetupClient(debug bool) *Setup {
	return &Setup{
		Debug: debug,
	}
}

func (s *Setup) execute() {
	s.createConfig()
	s.setupPlugins()
	s.setupCompletion()
}

func (s *Setup) createConfig() {
	var create bool

	cf := configfile.NewConfigFileClient(s.Debug)

	if cf.ConfigFileExists() {
		create = commandline.AskYesNoQuestion("It looks like the " + cf.ConfigFile() + " file already exists. Do you wish to replace it?")
	} else {
		create = true
	}

	if create {
		s.createOrUpdateConfig()
	}
}

func (s *Setup) createOrUpdateConfig() {
	content := s.generateConfigFileContents()
	cf := configfile.NewConfigFileClient(s.Debug)

	if !cf.ConfigDirExists() {
		err := os.Mkdir(cf.ConfigDir(), 0755)
		if err != nil {
			if s.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
			return
		}
	}

	err := os.WriteFile(cf.ConfigFile(), []byte(content), 0644)
	if err != nil {
		if s.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return
	}

	fmt.Printf("\nDone setting up %s!\n\n", cf.ConfigFile())
}

func (s *Setup) generateConfigFileContents() string {
	var contents string

	github := commandline.AskYesNoQuestion("Do you wish to set up GitHub credentials?")

	if github {
		contents = contents + "github_username: " + commandline.AskOpenEndedQuestion("GitHub username?", false) + "\n"
		contents = contents + "github_token: " + commandline.AskOpenEndedQuestion("GitHub personal access token? (Navigate to https://github.com/settings/tokens to create a new personal access token)", true) + "\n"
	}

	gitlab := commandline.AskYesNoQuestion("Do you wish to set up GitLab credentials?")

	if gitlab {
		contents = contents + "gitlab_username: " + commandline.AskOpenEndedQuestion("GitLab username?", false) + "\n"
		contents = contents + "gitlab_token: " + commandline.AskOpenEndedQuestion("GitLab personal access token? (Navigate to https://gitlab.com/-/profile/personal_access_tokens to create a new personal access token)", true) + "\n"
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
	cf := configfile.NewConfigFileClient(s.Debug)
	pluginsDir := cf.ConfigDir() + "/plugins"
	pluginsURL := "https://api.github.com/repos/emmahsax/go-git-helper/contents/plugins"

	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		if s.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return
	}

	resp, err := http.Get(pluginsURL)
	if err != nil {
		if s.Debug {
			debug.PrintStack()
		}
		log.Fatal("Error:", err)
		return
	}
	defer resp.Body.Close()

	var allPlugins []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allPlugins); err != nil {
		if s.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return
	}

	for _, plugin := range allPlugins {
		pluginURL := plugin["download_url"].(string)
		pluginName := plugin["name"].(string)

		resp, err := http.Get(pluginURL)
		if err != nil {
			if s.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
			continue
		}
		defer resp.Body.Close()

		pluginPath := filepath.Join(pluginsDir, pluginName)
		file, err := os.Create(pluginPath)
		if err != nil {
			if s.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
			continue
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			if s.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
			continue
		}

		err = os.Chmod(pluginPath, 0755)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	fmt.Printf("\nDone setting up plugins at %s!\n", pluginsDir)
	fmt.Printf("\nNow add this line to your ~/.zshrc:\n  export PATH=\"$HOME/.git_helper/plugins:$PATH\"\n")
}

func (s *Setup) setupCompletion() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up Git Helper completion?")

	if setup {
		shes := []string{"bash", "fish", "powershell", "zsh"}
		cf := configfile.NewConfigFileClient(s.Debug)
		completionsDir := cf.ConfigDir()+"/completions"
		if err := os.MkdirAll(completionsDir, 0755); err != nil {
			if s.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
			return
		}

		for _, sh := range shes {
			cmd := exec.Command("git-helper", "completion", sh)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if s.Debug {
					debug.PrintStack()
				}
				log.Fatal(err)
				return
			}

			filename := completionsDir+"/completion."+sh

			file, err := os.Create(filename)
			if err != nil {
				if s.Debug {
					debug.PrintStack()
				}
				log.Fatal(err)
				return
			}
			defer file.Close()

			_, err = file.WriteString(string(output))
			if err != nil {
				if s.Debug {
					debug.PrintStack()
				}
				log.Fatal(err)
				return
			}
		}

		fmt.Println("\nCompletions (for bash, fish, powershell, and zsh) generated in "+completionsDir+". Please activate the proper completion for your Unix shell.")
	}
}
