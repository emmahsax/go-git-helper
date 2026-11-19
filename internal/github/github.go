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
// PublicHost, or a self-hosted/Enterprise "github.example.com") and
// username. When more than one account matches host, username picks which
// one (see configfile.ResolveAccount); callers that already know there's no
// ambiguity can pass "".
func NewGitHub(host, username string, debugB bool) *GitHub {
	cf := configfile.NewConfigFile(debugB)
	account := configfile.ResolveAccount(cf.GitHubAccounts(host), username)

	c := newGitHubClient(account.Token, account.APIHost, debugB)

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

// newGitHubClient talks to a self-hosted GitHub Enterprise at apiHost (e.g.
// "github.example.com") when one is given, falling back to the SDK's
// default (the public github.com) for "". apiHost is the account's real API
// server (configfile.Account.APIHost) -- not whatever host matched the git
// remote, which may just be a local SSH alias with no meaning to any API.
func newGitHubClient(token, apiHost string, debugB bool) *github.Client {
	opts := []github.ClientOptionsFunc{github.WithAuthToken(token)}

	if apiHost != "" {
		baseURL := "https://" + apiHost + "/"
		opts = append(opts, github.WithEnterpriseURLs(baseURL, baseURL))
	}

	client, err := github.NewClient(opts...)
	if err != nil {
		utils.HandleError(err, debugB, nil)
	}
	return client
}
