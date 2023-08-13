# Git Helper (Go version)

Making it easier to work with Git on the command line.

## Installation & Setup

## Usage

### With Plugins

### With Aliases

## Commands

### `change-remote`

This can be used when switching the owners of a GitHub repo. When you switch a username, GitHub only makes some changes for you. With this command, you no longer have to manually walk through each local repo and switch the remotes from each one into a remote with the new username.

This command will go through every directory in a directory, and see if it is a git directory. It will then ask the user if they wish to process the git directory in question. The command does not yet know if there's any changes to be made. If the user says 'yes', then it will check to see if the old username is included in the remote URL of that git directory. If it is, then the command will change the remote URL to instead point to the new username's remote URL. To run the command, run:

```bash
git-helper change-remote [oldOwner] [newOwner]
```

### `checkout-default`

This command will check out the default branch of whatever repository you're currently in. It looks at what branch the `origin/HEAD` remote is pointed to on your local machine, versus querying GitHub/GitLab for that, so if your local machine's remotes aren't up to date or aren't formatted as expected, then this command won't work as expected. To run this command, run:

```bash
git-helper checkout-default
```

If your local branches formatted correctly (run `git branch --remote` to see), then run:

```bash
git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/CORRECT-DEFAULT-BRANCH-GOES-HERE
```

### `clean-branches`

This command will bring you to the repository's default branch, `git pull`, `git fetch -p`, and will clean up your local branches on your machine by seeing which ones are existing on the remote, and updating yours accordingly. To clean your local branches, run:

```bash
git-helper clean-branches
```

## Migrating from the Ruby version of Git Helper

---

### Contributing

To submit a feature request, bug ticket, etc, please submit an official [GitHub issue](https://github.com/emmahsax/go-git-helper/issues/new). To copy or make changes, please [fork this repository](https://github.com/emmahsax/go-git-helper/fork). When/if you'd like to contribute back to this upstream, please create a pull request on this repository.

Please follow included Issue Template(s) and Pull Request Template(s) when creating issues or pull requests.

### Security Policy

To report any security vulnerabilities, please view this repository's [Security Policy](https://github.com/emmahsax/go-git-helper/security/policy).

### Licensing

For information on licensing, please see [LICENSE.md](https://github.com/emmahsax/go-git-helper/blob/main/LICENSE.md).

### Code of Conduct

When interacting with this repository, please follow [Contributor Covenant's Code of Conduct](https://contributor-covenant.org).

### Releasing

To make a new release:

1. Merge the pull request via the big green button
2. Run `git tag vX.X.X` and `git push --tag`
3. A new release will be kicked off via GitHub Actions automatically
