package gitlab

import (
	"errors"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/utils"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

// PublicHost is the public GitLab's host. A provider entry in config.yml
// with no Hosts set is assumed to be this account.
const PublicHost = "gitlab.com"

type GitLab struct {
	Debug  bool
	Client *gitlab.Client
}

// NewGitLab builds a client for the GitLab account configured for host (e.g.
// PublicHost, or a self-hosted "gitlab.example.com") and username. When more
// than one account matches host, username picks which one (see
// configfile.ResolveAccount); callers that already know there's no
// ambiguity can pass "".
func NewGitLab(host, username string, debugB bool) *GitLab {
	cf := configfile.NewConfigFile(debugB)
	account := configfile.ResolveAccount(cf.GitLabAccounts(host), username)

	c, err := newGitLabClient(account.Token, account.APIHost, debugB)
	if err != nil {
		customErr := errors.New("could not create GitLab client: " + err.Error())
		utils.HandleError(customErr, debugB, nil)
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
		utils.HandleError(err, c.Debug, nil)
		return nil, err
	}

	return mr, nil
}

// newGitLabClient talks to a self-hosted GitLab at apiHost (e.g.
// "gitlab.example.com") when one is given, falling back to the SDK's default
// (the public gitlab.com) for "". apiHost is the account's real API server
// (configfile.Account.APIHost) -- not whatever host matched the git remote,
// which may just be a local SSH alias with no meaning to any API.
func newGitLabClient(token, apiHost string, debugB bool) (*gitlab.Client, error) {
	var opts []gitlab.ClientOptionFunc
	if apiHost != "" {
		opts = append(opts, gitlab.WithBaseURL("https://"+apiHost))
	}

	git, err := gitlab.NewClient(token, opts...)
	if err != nil {
		utils.HandleError(err, debugB, nil)
		return nil, err
	}
	return git, nil
}
