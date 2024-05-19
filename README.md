# Git Helper (Go version)

Making it easier to work with Git on the command-line.

## Installation

> All `task build`s and pre-built packages are built specifically using the flags `GOOS=darwin GOARCH=arm64`, only for Apple Silicon. To obtain any other build setups, either build your own or contact me at https://emmasax.com/contact-me/. I do not guarantee the performance of this package on any systems that are not Apple Silicon.

### Building Locally

You can build the binary locally by running this from the root directory of this repository:

```bash
task build
```

Then, you can move that to `/usr/local/bin`:

```bash
sudo mv git-helper_darwin_arm64 /usr/local/bin/git-helper
```

### Downloading from GitHub Releases

Alternatively, you can download pre-built binaries straight from [GitHub Releases](https://github.com/emmahsax/go-git-helper/releases). After downloading, move them into `/usr/local/bin` and change the permissions. The examples below assume the binaries downloaded to `~/Downloads`:

```bash
sudo mv ~/Downloads/git-helper_darwin_arm64 /usr/local/bin/git-helper
sudo chown $(whoami):staff /usr/local/bin/git-helper
sudo chmod +x /usr/local/bin/git-helper
```

When you run the binary for the first time, you may get a pop-up from Apple saying the binary can't be verified by Apple. If you wish to continue, follow these instructions to allow it to run:

* Click `Show in Finder`
* Right-click on the binary, and select `Open`
* When you get another pop-up, select `Open`

### Updating Git Helper

> This section assumes you already have Go's Git Helper installed.

To update your version of Git Helper to the latest version available in GitHub, you can run:

```bash
git-helper update
```

Your `sudo` password may be required, and you may need to verify the package is runable as following the instructions above.

## Config Setup

Some of the commands can be used without any additional configuration. However, others utilize special GitHub or GitLab configuration. To set up access with GitHub/GitLab, run:

```bash
git-helper setup
```

This will give you the option to set up credentials at GitHub and/or GitLab, as well as give you the choice to set up Git Helper as a plugin or not (see below).

The final result will be a `~/.git-helper/config.yml` file with the contents in this form:

```
github_user:  GITHUB-USERNAME
github_token: GITHUB-TOKEN
gitlab_user:  GITLAB-USERNAME
gitlab_token: GITLAB-TOKEN
```

To create or see what personal access tokens (PATs) you have, look [here for GitHub PATs](https://github.com/settings/tokens) and [here for GitLab PATs](https://gitlab.com/-/profile/personal_access_tokens). You could either have one set of tokens for each computer you use, or just have one set of tokens for all computers that you rotate periodically.

## General Usage

In general, all commands can be run straight from the command-line like this:

```bash
git-helper [command]
```

Please run with `help` to see more:

```bash
git-helper --help
```

In addition, hopefully all these options below make working with git and Go's Git Helper more seamless.

### With Plugins

As an additional enhancement, you can set each of the following commands to be a git plugin, meaning you can call them in a way that feels more git-native:

```bash
# Without plugins
git-helper clean-branches
git-helper code-request

# With plugins
git clean-branches
git code-request
```

Running the `git-helper setup` command will give you the option to set plugins up.

### With Aliases

To make the commands even shorter, I recommend setting aliases. You can either set aliases through git itself, like this (only possible if you also use the plugin option):

```bash
git config --global alias.nb new-branch
```

And then running `git nb` maps to `git new-branch`, which through the plugin, maps to `git-helper new-branch`.

Or you can set the alias through your `~/.zshrc` (which is my preferred method because it can make the command even shorter, and doesn't require the plugin usage). To do this, add lines like this to the `~/.zshrc` file and run `source ~/.zshrc`:

```bash
alias gnb="git new-branch"
```

And then, running `gnb` maps to `git new-branch`, which again routes to `git-helper new-branch`.

For a full list of the git aliases I prefer to use, check out my [Git Aliases gist](https://gist.github.com/emmahsax/e8744fe253fba1f00a621c01a2bf68f5).

### With Completion

To setup completion (auto-filling commands as you type), run:

```bash
mkdir -p ~/.git-helper/completions
git-helper completion bash >> ~/.git-helper/completions/completion.bash
git-helper completion fish >> ~/.git-helper/completions/completion.fish
git-helper completion powershell >> ~/.git-helper/completions/completion.powershell
git-helper completion zsh >> ~/.git-helper/completions/completion.zsh
```

Then load the appropriate completion file for your scripting language to this to your shell file (e.g. load the following into your `~/.zshrc`):

```bash
source ~/.git-helper/completions/completion.zsh
```

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

### `clean-branches`

This command will bring you to the repository's default branch, `git pull`, `git fetch -p`, and will clean up your local branches on your machine by seeing which ones are existing on the remote, and updating yours accordingly. To clean your local branches, run:

```bash
git-helper clean-branches
```

### `code-request`

This command can be used to handily make new GitHub/GitLab pull/merge requests from the command-line. The command uses either the [GitHub REST API](https://docs.github.com/en/rest) or [GitLab API](https://docs.gitlab.com/ee/api/) to do this, so make sure you have a `~/.git-helper/config.yml` file set up in the home directory of your computer (instructions are higher in this `README`).

After setup is complete, you can call the command like this:

```bash
git-helper code-request
```

The command will provide an autogenerated code request title based on your branch name. It will separate the branch name by `'_'` if underscores are in the branch, or `'-'` if dashes are present. Then it will join the list of words together by spaces. If there's a pattern in the form of `jira-123` or `jira_123` in the first part of the branch name, then it'll add `JIRA-123` to the first part of the code request. You can choose whether to accept this title or not. If the title's declined, you can provide your own code request title.

The command will also ask you if the default branch of the repository is the proper base branch to use. You can say whether that is correct or not, and if it's incorrect, you can give the command your chosen base branch.

The command will also ask you whether you'd like to mark the new code request as a draft or not.

If your project uses GitLab, the command will automatically set the merge request to delete the source branch upon merge. The value can later be changed for a specific MR either over the API or in the browser. The command also automatically sets the merge request to squash, and this will be the setting on the MR if the project allows, encourages, or requires squashing. If the project doesn't allow squashing at all, then that option will be voided, and the MR will not be squashed. Depending on the project's settings, the value can later be changed for a specific MR over the API or in the browser.

Then, it'll ask about code request templates. For GitHub, it'll ask the user to apply any pull request templates found at `.github/pull_request_template.md`, `./pull_request_template.md`, or `.github/PULL_REQUEST_TEMPLATE/*.md`. Applying any template is optional, and a user can make an empty pull request if they desire. For GitLab, it'll ask the user to apply any merge request templates found at any `.gitlab/merge_request_template.md`, `./merge_request_template.md`, or `.gitlab/merge_request_templates/*.md`. Applying any template is optional, and from the command's standpoint, a user can make an empty merge request if they desire (although GitLab may still add a merge request template if the project itself requires one). When searching for templates, the code ignores cases, so the file could be named with all capital letters or all lowercase letters.

### `empty-commit`

For some reason, I'm always forgetting the commands to create an empty commit. So with this command, it becomes easy. The commit message of this commit will be `Empty commit`. To run the command, run:

```bash
git-helper empty-commit
```

### `forget-local-changes`

This command will quickly and easily get rid of any local changes that are not in a commit. This command does just a `git stash` and `git stash drop`. **Once you forget them, they're completely gone, so run carefully**. To test it out, run:

```bash
git-helper forget-local-changes
```

### `forget-local-commits`

This command is handy if you locally have a bunch of commits you wish to completely get rid of. This command basically does a hard reset to `origin/HEAD`. **Once you forget them, they're completely gone, so run carefully**. To test it out, run:

```bash
git-helper forget-local-commits
```

### `new-branch`

This command is useful for making new branches in a repository on the command-line. To run the command, run:

```bash
git-helper new-branch
# OR
git-helper new-branch [optionalBranch]
```

The command either accepts a branch name right away or it will ask you for the name of your new branch. Make sure your input does not contain any spaces or special characters.

### `set-head-ref`

Sets the upstream and `HEAD` symbolic ref to the default branch passed in:

```bash
git-helper set-head-ref [defaultBranch]
```

### `setup`

See [`Setup`](#setup).

### `update`

See [`Updating Git Helper`](#updating-git-helper).

### `version`

Show's the local version of Git Helper installed:

```bash
git-helper version
```

## Migrating from the Ruby version of Git Helper

1. Uninstall Ruby's Git Helper:
    ```bash
    # Uninstall the gem
    gem uninstall git_helper

    # Remove the executable
    which git-helper
    rm -rf PATH/FROM/ABOVE

    # Verify it's gone by this command returning command not found
    git-helper -v
    ```
2. Install Go's Git Helper by following the instructions in [`Installation`](#installation)
3. Check Go's Git Helper is installed:
    ```bash
    git-helper version
    ```
4. Run the setup command (optional for Beta modes < 1.0.0, required otherwise):
    ```bash
    git-helper setup
    ```
5. Move forward with your day!

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

1. Verify `main` has or will have the newest version in the `main.go` file
1. Merge the pull request via the big green button
3. Trigger a new workflow from [GitHub Actions](https://github.com/emmahsax/go-git-helper/actions/workflows/release.yml) and pass in the package version indicated in the `main.go` file (but include the `v` prefix)
