package github

import (
	"context"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

type GitHub struct {
	Debug  bool
	Client *github.Client
}

func NewGitHub(debugB bool) *GitHub {
	cf := configfile.NewConfigFile(debugB)
	c := newGitHubClient(cf.GitHubToken())

	return &GitHub{
		Debug:  debugB,
		Client: c,
	}
}

func (c *GitHub) CreatePullRequest(owner, repo string, options *github.NewPullRequest) (*github.PullRequest, error) {
	pr, _, err := c.Client.PullRequests.Create(context.Background(), owner, repo, options)
	if err != nil {
		utils.HandleError(err, c.Debug, nil)
		return nil, err
	}

	return pr, nil
}

func newGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	git := github.NewClient(tc)
	return git
}
