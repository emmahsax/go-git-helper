package gitlab

import (
	"errors"

	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/utils"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitLab struct {
	Debug  bool
	Client *gitlab.Client
}

func NewGitLab(debugB bool) *GitLab {
	cf := configfile.NewConfigFile(debugB)
	c, err := newGitLabClient(cf.GitLabToken(), debugB)
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

func newGitLabClient(token string, debugB bool) (*gitlab.Client, error) {
	git, err := gitlab.NewClient(token)
	if err != nil {
		utils.HandleError(err, debugB, nil)
		return nil, err
	}
	return git, nil
}
