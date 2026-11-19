package codeRequest

import (
	"github.com/emmahsax/go-git-helper/internal/github"
	"github.com/emmahsax/go-git-helper/internal/utils"
)

func (cr *CodeRequest) createGitHub() {
	host := cr.matchedGitHubHost()

	options := make(map[string]string)
	options["baseBranch"] = cr.baseBranch()
	options["draft"] = cr.draft()
	options["host"] = host
	options["username"] = cr.chosenUsername("GitHub", cr.config().Accounts("github", host))
	options["newPrTitle"] = cr.newPrTitle()
	g := cr.git()
	options["gitRootDir"] = g.GetGitRootDir()
	options["localBranch"] = g.CurrentBranch()
	options["localRepo"] = g.RepoName()
	options["ticketBaseURLCandidates"] = utils.EncodeTicketBaseURLCandidates(cr.ticketBaseURLCandidates(options["newPrTitle"], options["localRepo"]))
	github.NewGitHubPullRequest(options, cr.Debug, cr.InteractiveMode).Create()
}

func (cr *CodeRequest) isGitHub() bool {
	return cr.matchedGitHubHost() != ""
}

// matchedGitHubHost returns whichever of the public github.com or a
// self-hosted/Enterprise host from the "github" entries in config.yml
// matches one of the repo's git remote hosts, or "" if none do. See
// matchedHost for the exact matching rules.
func (cr *CodeRequest) matchedGitHubHost() string {
	return matchedHost(cr.git().RemoteHosts(), github.PublicHost, cr.config().Hosts("github"))
}
