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
// PublicHost, or a self-hosted "gitlab.example.com"), picking whichever of
// the config file's (possibly several) "gitlab" provider entries matches.
func NewGitLab(host string, debugB bool) *GitLab {
	cf := configfile.NewConfigFile(debugB)
	_, token := cf.GitLabCredentials(host)

	c, err := newGitLabClient(token, host, debugB)
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

// newGitLabClient talks to a self-hosted GitLab at host (e.g.
// "gitlab.example.com") when one is given, falling back to the SDK's default
// (the public gitlab.com) for "" or PublicHost.
func newGitLabClient(token, host string, debugB bool) (*gitlab.Client, error) {
	var opts []gitlab.ClientOptionFunc
	if host != "" && host != PublicHost {
		opts = append(opts, gitlab.WithBaseURL("https://"+host))
	}

	git, err := gitlab.NewClient(token, opts...)
	if err != nil {
		utils.HandleError(err, debugB, nil)
		return nil, err
	}
	return git, nil
}
