package codeRequest

import "github.com/emmahsax/go-git-helper/internal/gitlab"

func (cr *CodeRequest) createGitLab() {
	host := cr.matchedGitLabHost()

	options := make(map[string]string)
	options["baseBranch"] = cr.baseBranch()
	options["draft"] = cr.draft()
	options["host"] = host
	options["username"] = cr.chosenUsername("GitLab", cr.config().Accounts("gitlab", host))
	options["newMrTitle"] = cr.newMrTitle()
	g := cr.git()
	options["gitRootDir"] = g.GetGitRootDir()
	options["localBranch"] = g.CurrentBranch()
	options["localProject"] = g.RepoName()
	options["ticketBaseURL"] = cr.ticketBaseURL(options["newMrTitle"], options["localProject"])
	gitlab.NewGitLabMergeRequest(options, cr.Debug, cr.InteractiveMode).Create()
}

func (cr *CodeRequest) isGitLab() bool {
	return cr.matchedGitLabHost() != ""
}

// matchedGitLabHost returns whichever of the public gitlab.com or a
// self-hosted host from the "gitlab" entries in config.yml matches one of
// the repo's git remote hosts, or "" if none do. See matchedHost for the
// exact matching rules.
func (cr *CodeRequest) matchedGitLabHost() string {
	return matchedHost(cr.git().RemoteHosts(), gitlab.PublicHost, cr.config().Hosts("gitlab"))
}
