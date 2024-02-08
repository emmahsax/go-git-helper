package gitlab

import (
	"log"
	"runtime/debug"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/xanzy/go-gitlab"
)

type GitLab struct {
	Debug  bool
	Client *gitlab.Client
}

func NewGitLab(debugB bool) *GitLab {
	cf := configfile.NewConfigFile(debugB)
	c, err := newGitLabClient(cf.GitLabToken())
	if err != nil {
		if debugB {
			debug.PrintStack()
		}
		log.Fatal("Could not create GitLab client: ", err)
		return nil
	}

	return &GitLab{
		Debug:  debugB,
		Client: c,
	}
}

func (c *GitLab) CreateMergeRequest(projectName string, options *gitlab.CreateMergeRequestOptions) (*gitlab.MergeRequest, error) {
	mr, _, err := c.Client.MergeRequests.CreateMergeRequest(projectName, options)
	if err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return nil, err
	}

	return mr, nil
}

func newGitLabClient(token string) (*gitlab.Client, error) {
	git, err := gitlab.NewClient(token)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return git, nil
}
