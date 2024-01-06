package github

import (
	"context"
	"log"
	"runtime/debug"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/google/go-github/v56/github"
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

func (c *GitHub) CreatePullRequest(owner, repo string, options map[string]string) (*github.PullRequest, error) {
	createOpts := &github.NewPullRequest{
		Base:                github.String(options["target_branch"]),
		Body:                github.String(options["description"]),
		Head:                github.String(options["source_branch"]),
		MaintainerCanModify: github.Bool(true),
		Title:               github.String(options["title"]),
	}

	pr, _, err := c.Client.PullRequests.Create(context.Background(), owner, repo, createOpts)
	if err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
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
