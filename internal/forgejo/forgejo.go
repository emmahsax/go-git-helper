package forgejo

import (
	"errors"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v4"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/utils"
)

// PublicHost is Codeberg, the most prominent public Forgejo instance. A
// provider entry in config.yml with no RemoteHosts set is assumed to be this
// account. Unlike GitHub/GitLab, Forgejo has no single centralized SaaS --
// most accounts are genuinely self-hosted, so most entries will set both
// RemoteHosts and APIHost.
const PublicHost = "codeberg.org"

type Forgejo struct {
	Debug  bool
	Client *forgejo.Client
}

// NewForgejo builds a client for the Forgejo account configured for host
// (e.g. PublicHost, or a self-hosted "forgejo.example.com") and username.
// When more than one account matches host, username picks which one (see
// configfile.ResolveAccount); callers that already know there's no ambiguity
// can pass "".
func NewForgejo(host, username string, debugB bool) *Forgejo {
	cf := configfile.NewConfigFile(debugB)
	account := configfile.ResolveAccount(cf.Accounts("forgejo", host), username)

	c, err := newForgejoClient(account.Token, account.APIHost, debugB)
	if err != nil {
		customErr := errors.New("could not create Forgejo client: " + err.Error())
		utils.HandleError(customErr, debugB, nil)
		return nil
	}

	return &Forgejo{
		Debug:  debugB,
		Client: c,
	}
}

func (c *Forgejo) CreatePullRequest(owner, repo string, options forgejo.CreatePullRequestOption) (*forgejo.PullRequest, error) {
	pr, _, err := c.Client.CreatePullRequest(owner, repo, options)
	if err != nil {
		utils.HandleError(err, c.Debug, nil)
		return nil, err
	}

	return pr, nil
}

// newForgejoClient talks to the Forgejo instance at apiHost (e.g.
// "forgejo.example.com") when one is given, falling back to PublicHost
// (Codeberg) for "". Unlike the GitHub/GitLab SDKs, forgejo-sdk has no
// built-in default API server, so a base URL is always passed explicitly.
// apiHost is the account's real API server (configfile.Account.APIHost) --
// not whatever host matched the git remote, which may just be a local SSH
// alias with no meaning to any API.
//
// forgejo.NewClient otherwise makes a live GET /api/v1/version call against
// the server (with the token already attached) before returning, purely to
// gate version-dependent SDK features. CreatePullRequestOption uses none of
// those, so SetForgejoVersion("") skips that call -- an unreachable or slow
// self-hosted instance shouldn't fail client construction before the real
// request (CreatePullRequest) even gets a chance to run and report the
// actual problem.
func newForgejoClient(token, apiHost string, debugB bool) (*forgejo.Client, error) {
	host := apiHost
	if host == "" {
		host = PublicHost
	}

	client, err := forgejo.NewClient("https://"+host, forgejo.SetToken(token), forgejo.SetForgejoVersion(""))
	if err != nil {
		utils.HandleError(err, debugB, nil)
		return nil, err
	}

	return client, nil
}
