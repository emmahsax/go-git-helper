package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/google/go-github/v88/github"
)

type GitHub struct {
	Debug  bool
	Client *github.Client
}

func NewGitHub(debugB bool) *GitHub {
	cf := configfile.NewConfigFile(debugB)
	c := newGitHubClient(cf.GitHubToken(), debugB)

	return &GitHub{
		Debug:  debugB,
		Client: c,
	}
}

func (c *GitHub) CreatePullRequest(owner, repo string, options *github.NewPullRequest) (*github.PullRequest, error) {
	var err error
	var pr *github.PullRequest

	for {
		pr, _, err = c.Client.PullRequests.Create(context.Background(), owner, repo, options)
		if err != nil {
			if strings.Contains(err.Error(), "422 Draft pull requests are not supported in this repository.") {
				fmt.Println("Draft pull requests are not supported in this repository. Retrying.")
				options.Draft = github.Ptr(false)
				continue
			}
			utils.HandleError(err, c.Debug, nil)
			return nil, err
		}

		break
	}

	return pr, nil
}

func newGitHubClient(token string, debugB bool) *github.Client {
	client, err := github.NewClient(github.WithAuthToken(token))
	if err != nil {
		utils.HandleError(err, debugB, nil)
	}
	return client
}
