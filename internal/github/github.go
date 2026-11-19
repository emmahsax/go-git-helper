package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/google/go-github/v91/github"
)

// PublicHost is the public GitHub's host. A provider entry in config.yml
// with no Hosts set is assumed to be this account.
const PublicHost = "github.com"

type GitHub struct {
	Debug  bool
	Client *github.Client
}

// NewGitHub builds a client for the GitHub account configured for host (e.g.
// PublicHost, or a self-hosted/Enterprise "github.example.com"), picking
// whichever of the config file's (possibly several) "github" provider
// entries matches.
func NewGitHub(host string, debugB bool) *GitHub {
	cf := configfile.NewConfigFile(debugB)
	_, token := cf.GitHubCredentials(host)

	c := newGitHubClient(token, host, debugB)

	return &GitHub{
		Debug:  debugB,
		Client: c,
	}
}

func (c *GitHub) CreatePullRequest(owner, repo string, options *github.CreatePullRequest) (*github.PullRequest, error) {
	var err error
	var pr *github.PullRequest

	for {
		pr, _, err = c.Client.PullRequests.Create(context.Background(), owner, repo, *options)
		if err != nil {
			if strings.Contains(err.Error(), "422 Draft pull requests are not supported in this repository.") {
				fmt.Println("Draft pull requests are not supported in this repository. Retrying...")
				options.Draft = new(false)
				continue
			}
			utils.HandleError(err, c.Debug, nil)
			return nil, err
		}

		break
	}

	return pr, nil
}

// newGitHubClient talks to a self-hosted GitHub Enterprise at host (e.g.
// "github.example.com") when one is given, falling back to the SDK's
// default (the public github.com) for "" or PublicHost.
func newGitHubClient(token, host string, debugB bool) *github.Client {
	opts := []github.ClientOptionsFunc{github.WithAuthToken(token)}

	if host != "" && host != PublicHost {
		baseURL := "https://" + host + "/"
		opts = append(opts, github.WithEnterpriseURLs(baseURL, baseURL))
	}

	client, err := github.NewClient(opts...)
	if err != nil {
		utils.HandleError(err, debugB, nil)
	}
	return client
}
