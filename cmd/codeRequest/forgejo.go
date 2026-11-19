package codeRequest

import "github.com/emmahsax/go-git-helper/internal/forgejo"

func (cr *CodeRequest) createForgejo() {
	host := cr.matchedForgejoHost()

	options := make(map[string]string)
	options["baseBranch"] = cr.baseBranch()
	options["draft"] = cr.draft()
	options["host"] = host
	options["username"] = cr.chosenUsername("Forgejo", cr.config().Accounts("forgejo", host))
	options["newPrTitle"] = cr.newPrTitle()
	g := cr.git()
	options["gitRootDir"] = g.GetGitRootDir()
	options["localBranch"] = g.CurrentBranch()
	options["localRepo"] = g.RepoName()
	options["ticketBaseURL"] = cr.ticketBaseURL(options["newPrTitle"], options["localRepo"])
	forgejo.NewForgejoPullRequest(options, cr.Debug, cr.InteractiveMode).Create()
}

func (cr *CodeRequest) isForgejo() bool {
	return cr.matchedForgejoHost() != ""
}

// matchedForgejoHost returns whichever of the public codeberg.org or a
// self-hosted host from the "forgejo" entries in config.yml matches one of
// the repo's git remote hosts, or "" if none do. See matchedHost for the
// exact matching rules.
func (cr *CodeRequest) matchedForgejoHost() string {
	return matchedHost(cr.git().RemoteHosts(), forgejo.PublicHost, cr.config().Hosts("forgejo"))
}
