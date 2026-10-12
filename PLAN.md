# Cross-Platform Plan: macOS Intel, Linux, Windows

## 1. Summary

The core is already close to portable: pure Go, no cgo, and all git work goes through `exec.Command("git", ...)`. `filepath` is used for most path joins. The blockers are in packaging and in a few commands that assume a Unix install layout. That means `update`, `setup` (plugins and completions), the build and release pipeline, and the docs.

Everything builds for other targets today with `GOOS=<os> GOARCH=<arch> go build`. Nothing currently produces, publishes or tests those builds.

**Target matrix**

| Target | GOOS/GOARCH | Today |
|---|---|---|
| macOS Apple Silicon | darwin/arm64 | Only supported build |
| macOS Intel | darwin/amd64 | Not built or released. Code is fine. |
| Linux x86_64 | linux/amd64 | Not built. `update` partly handles it. |
| Linux ARM64 (optional) | linux/arm64 | Same as Linux x86_64 |
| Windows x86_64 | windows/amd64 | Not built. `update` and `setup` break. |
| Windows ARM64 (optional) | windows/arm64 | Same as Windows x86_64 |

## 2. Findings

### 2.1 Build and release (blocking for all new targets)

- `Taskfile.yml`: `build` calls only `build:darwin:arm64`. The `build:test` task also hardcodes darwin/arm64, uses `sudo mv` to `/usr/local/bin`, and has no Windows equivalent.
- `.github/workflows/release.yml`: `files:` lists only `git-helper_darwin_arm64`, so no other asset is uploaded.
- `.github/workflows/test.yml`: it runs `task build` and the tests on `ubuntu-latest` only (through the shared `go-test.yml` workflow). There are no macOS or Windows runners, so a platform regression would not be caught.
- The asset naming convention `git-helper_<GOOS>_<GOARCH>` is used by `cmd/update/update.go:26`. The release assets must follow it exactly. Windows assets need the `.exe` suffix, so `update` has to account for it.
- No checksums or signatures are published.

### 2.2 `cmd/update/update.go` (functional bugs on non-arm64 and Windows)

- `newPath = "/usr/local/bin/git-helper"` (line 27) is hardcoded. It is wrong for Windows. It is also wrong for Apple Silicon Homebrew (`/opt/homebrew/bin`), for Linux installs in `~/.local/bin`, and wherever the user actually put the binary.
- `getDownloadURL` has a `darwin` and a `linux` case with identical bodies, and a `default` that prints "Unsupported" but then continues with an empty URL. There is no `windows` case.
- `moveGitHelper` and `setPermissions` shell out to `sudo mv`, `sudo chown user:staff` and `sudo chmod +x`. These have several problems:
  - `sudo` does not exist on Windows.
  - The `staff` group does not exist on most Linux distros (the usual primary group is the user's own name), so `chown` fails there.
  - Requiring `sudo` is unnecessary when the install directory is user-writable.
- The release asset is downloaded into the current working directory (`os.Create(binaryName)`), not a temp dir. It leaves files behind and can overwrite a local file of the same name.
- `http.Get` has no status-code check. A 404 or a rate-limit JSON body gets written out as the "binary".
- A running `.exe` cannot be overwritten on Windows, so an in-place update needs a rename-then-replace approach.
- `outputNewVersion` runs `git-helper version` from `PATH`. That can report a different copy than the one just installed.

### 2.3 `cmd/setup/setup.go` (plugins and completions)

- **Plugins** (`plugins/git-*`) are `#!/bin/sh` scripts that call `git-helper <cmd> $@`. They are downloaded from the GitHub contents API into `~/.git-helper/plugins`, then `chmod 0755`.
  - On Windows these shell scripts do not run natively. Git for Windows ships `sh.exe`, so extensionless scripts on `PATH` can work under Git Bash but not from `cmd` or PowerShell.
  - The scripts use unquoted `$@` instead of `"$@"`, which breaks arguments with spaces on every OS. This is a small bug fix.
  - The printed instruction is `export PATH="$HOME/.git-helper/plugins:$PATH"` and mentions `~/.zshrc`. It is Unix-only and zsh-centric.
- **Completions** are generated for bash, fish, powershell and zsh, which is good. The activation hint is zsh-only and the generated file is named `completion.<sh>` (`completion.powershell` is fine). The instructions for PowerShell and bash on Linux are missing.
- `os.Mkdir` (config dir) and `os.MkdirAll` use `0755` and files use `0644`. On Windows the mode bits are mostly ignored, so this is harmless, but secrets in `config.yml` end up world-readable on Unix. **Recommendation:** write `config.yml` with `0600` and the dir with `0700`. This matters most on shared Linux machines.
- Hardcoded `"/plugins"`, `"/completions"` and `"/.git-helper"` string concatenation (`setup.go:129,195,208`, `configfile.go:40,55`). This is fine on Windows (Go accepts `/`) but inconsistent. Use `filepath.Join`.

### 2.4 `internal/configfile/configfile.go`

- `os.UserHomeDir()` is portable. On Windows it resolves `%USERPROFILE%`, so config lands in `C:\Users\<name>\.git-helper`.
- Decision needed: keep `~/.git-helper` everywhere (simplest, and existing users keep their config), or use OS-native dirs (`%AppData%`, `$XDG_CONFIG_HOME`). **Recommendation:** keep `~/.git-helper` for now. Optionally add `XDG_CONFIG_HOME` and `%AppData%` support later, with a fallback to the legacy dir.
- `ConfigDirExists()` has a latent nil-pointer bug. If `os.Stat` fails with an error other than not-exist, it falls through and calls `info.IsDir()` on a nil `info`. Fix this while in here.

### 2.5 Tests

- Tests that set `HOME` (`internal/configfile/configfile_test.go:132,237`) will not isolate on Windows, where `os.UserHomeDir()` reads `USERPROFILE`. Use `t.Setenv` for both `HOME` and `USERPROFILE`.
- `internal/gitlab/gitlab_test.go:15` and the `codeRequest` tests read the real `$HOME/.git-helper`, which is not hermetic.
- Path assertions that use `"/"` concatenation, such as `setup_test.go:40,219,225,248` and `configfile_test.go:38,77`, may mismatch on Windows if the production code moves to `filepath.Join`.
- `update_test.go:86,114` assert `sudo`-style args and the `/usr/local/bin` path. These will change with the `update` redesign.
- Tests that spawn a real git or a shell need `git` on `PATH` in each CI runner. Windows runners have Git for Windows preinstalled.

### 2.6 Other code

- `internal/executor/executor.go`: `exec.Command("git", ...)` works on every OS as long as `git` is on `PATH`. The `waitAndStdout` function reassigns `os.Stdout` and `os.Stderr` to themselves, which is a no-op. Not a portability issue, just dead code.
- Output parsing assumes `\n` line endings from git. Git emits `\n` on Windows too, but parsing in `internal/git/git.go` (`branch`, `branch -vv`, `remote -v`) should `TrimSpace` each line to tolerate a stray `\r`.
- `cmd/changeRemote/changeRemote.go:69` uses `filepath.Join(current, ".git")` and handles it correctly.
- `codeRequest.go:209,214` treats both `/` and `\` as separators in branch-name heuristics, so it is already Windows-aware.
- Interactive prompts use `pterm` (with `atomicgo/keyboard`, `containerd/console`, `golang.org/x/sys`). It supports Windows terminals, but interactive select and masked input should be verified manually in `cmd.exe`, Windows Terminal and PowerShell.
- Dependencies (go-github, GitLab client, cobra, yaml, gjson) are pure Go, with no OS-specific issues.
- Root `.gitignore` already ignores `*.exe`. Add the new build output pattern (see 3.1).

### 2.7 Docs

`README.md` says only Apple Silicon is supported. Install steps assume macOS (`sudo mv` to `/usr/local/bin`, `chown ...:staff`, Gatekeeper pop-up instructions, `~/.zshrc` aliases). It needs sections for macOS Intel, Linux and Windows.

## 3. Plan of Work

### Phase 1: Build matrix and release (unblocks everything else)

1. Rewrite the `Taskfile.yml` build tasks:
   - `build:darwin:arm64`, `build:darwin:amd64`, `build:linux:amd64`, `build:linux:arm64`, `build:windows:amd64` and `build:windows:arm64`. Windows outputs end in `.exe`, for example `git-helper_windows_amd64.exe`.
   - `CGO_ENABLED=0` on all of them for static, reproducible binaries.
   - A `build` task that runs all of them. Use `-ldflags "-s -w"`. Do not inject the version through ldflags (see the note under Phase 1).
   - Make `build:test` use the host OS and arch (`go env GOOS GOARCH`) and avoid `sudo` on Windows.
2. Update `release.yml` so `files:` lists every produced asset, plus a `checksums.txt` (SHA-256). But COMMENT THOSE LINES OUT FOR NOW TO SHOW THE CODE CAN BUILD THERE, BUT WE ARE NOT ACTIVELY PUSHING THOSE FILES.
3. Check `emmahsax/github-actions` `go-release.yml` to see whether it supports multi-file uploads and a matrix. Adjust it there if not.
4. The shared `go-test.yml` is hardcoded to `runs-on: ubuntu-latest` and has no OS input. It also runs `go vet`, `gofmt` and `staticcheck`. To test on other OSes, add a `runs-on` input (default `ubuntu-latest`) to `go-test.yml` in `emmahsax/github-actions`, then call it from `test.yml` with a matrix of `ubuntu-latest`, `ubuntu-24.04-arm`, `macos-latest` (arm64), `macos-15-intel` (Intel), `windows-latest` and `windows-11-arm`. Existing callers keep working because of the default. Run `task build` (all six targets) on every matrix leg or once on Linux as a cross-compile smoke test. The `gofmt-check` action and `setup-go` may also need checks for Windows (shell syntax), which I haven't read. If changes are required to `go-test.yml`, then make building on other runner types optional based on input values (where only `ubuntu-latest` is default to `true`).
5. Add the new binary patterns to `.gitignore`.

**Note on versioning:** the local `version-bump-check.yml` workflow requires every PR to change `packageVersion` in `main.go` (regex `\s+packageVersion\s+=\s+"\S+"`). The version lives in source and is bumped per PR, so tag-based ldflags injection would fight that workflow. Keep the current approach.

#### Phase 1 option: GoReleaser (for consolidating all three Go projects)

This is an alternative way to do Phase 1 items 1–3. Items 4 and 5 (test matrix and `.gitignore`) are the same either way. Your own point stands: the benefit is one build and release definition shared across three repos, not new features for this repo alone.

**Shape**
- Add `.goreleaser.yaml` to each repo. The per-repo parts are the project name and `main` path. The target matrix, `CGO_ENABLED=0`, `-s -w` and the naming template are the same everywhere.
- Set `archives: [{format: binary}]` and `name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"`, so assets are bare binaries named `git-helper_<os>_<arch>`. That is what `update` expects. GoReleaser should add `.exe` for Windows itself, but verify this with a snapshot build before relying on it. Use `{{ .Arch }}` (raw `amd64`), not the default template's `x86_64`.
- `checksum: {name_template: checksums.txt}` produces the checksums file.
- Keep one shared workflow, `go-release.yml`, and make GoReleaser optional inside it. Repos that don't opt in see no change, and no workflow file is renamed or replaced later.
  - Add a boolean input `use_goreleaser` (default `false`) and an optional `goreleaser_args` string (default `release --clean`).
  - Put `if: ${{ !inputs.use_goreleaser }}` on the existing `Go Build`, `GitHub Release & Upload Assets` and `GitHub Release` steps.
  - Add one step with `if: ${{ inputs.use_goreleaser }}` that runs `goreleaser/goreleaser-action` with `args: ${{ inputs.goreleaser_args }}` and `GITHUB_TOKEN`. The existing `Set up Go` step already runs first.
  - `go-release.yml` currently checks out `ref: main` with the default shallow depth. GoReleaser needs full history for the changelog and tag detection, so set `fetch-depth: 0` (harmless for the other path).
  - In this repo, `release.yml` then only adds `use_goreleaser: true`. When the other two repos move, they add the same line and keep their `release.yml` otherwise unchanged.
- Use `goreleaser build --snapshot --clean` (or `release --snapshot`) in the test workflow as the cross-compile smoke test. This is the same command a developer runs locally.
- Keep the "commented out" requirement: the release config can list all targets while the workflow's `upload` stays limited to `darwin_arm64`. GoReleaser can do this by building all targets but filtering which artifacts upload, or by leaving the other targets in the config behind a build `ids` filter. Which of those is cleaner needs a quick trial.

**Decisions on the open items**
1. **Tags: create the tag as the first step (agreed).** Add this to `go-release.yml`, before the GoReleaser step and gated on `use_goreleaser`:
   - Check out `main` with `fetch-depth: 0`.
   - If `tag_name` doesn't exist, create it on the checked-out commit and push it with the workflow's `GITHUB_TOKEN` (`contents: write` is already granted).
   - If it exists and points at `HEAD`, continue, so a failed run can be retried. If it exists on a different commit, fail with a clear message.
   - GoReleaser then finds the tag on `HEAD` with no special flags. (Pushes made with `GITHUB_TOKEN` don't trigger other workflows, which is fine here.)
   - Your tag styles differ: `go-git-helper` uses `1.0.0`-style, and `escli` uses plain integers like `46`. GoReleaser prefers semver tags, and I believe it warns on others but I haven't confirmed that it still proceeds. Test a snapshot or a throwaway release on `escli` before migrating it.
2. **Remove the version from source (agreed).**
   - Delete `packageVersion` from `main.go` and delete `.github/workflows/version-bump-check.yml`. The shared `version-bump-check.yml` in `emmahsax/github-actions` can stay unused.
   - `git-helper version` and `escli version` still need a version. Declare `var packageVersion = "dev"` (or `version`) in `main.go` and set it with `ldflags: -X main.packageVersion={{ .Version }}` in `.goreleaser.yaml`. As a fallback for `go install` builds, read `runtime/debug.ReadBuildInfo().Main.Version`. The `version` command already takes the version as a parameter, so its tests don't change.
   - Local `task build` builds report `dev`, unless the Taskfile passes `-ldflags` too (for example from `git describe --tags`).
   - `update` should compare the installed version with the latest release. With version `dev` it should always offer the update.
   - Drop the tag-must-match-`packageVersion` check I suggested earlier, since there is nothing left to match.
3. **Keep the GitHub release notes: yes.** Set `changelog.use: github-native` in `.goreleaser.yaml`, which asks GitHub to generate the notes, the same as `generate_release_notes: true` does today. Verify the key name against the current docs. The fallback is `changelog.disable: true` plus letting `action-gh-release` create the release, but I don't expect to need it.
4. **The other two projects (read).**
   - `~/git/escli`: the same shape as `go-git-helper`. Its `Taskfile.yml`, `release.yml` and `version-bump-check.yml` are near-identical, and it has `packageVersion = "46"` in `main.go`. It builds only `escli_darwin_arm64`. It can reuse the same `.goreleaser.yaml` with the project name changed, and the same version removal.
   - `~/git/github-pr-slack-notifier`: not a CLI. It builds an AWS Lambda package, `dist/lambda.zip`, from `./cmd/lambda` for `linux/arm64` with `CGO_ENABLED=0 -trimpath`, containing `bootstrap` and `LICENSE.md`. It has no version in source and no version-bump workflow.
     - GoReleaser can do it with a single build (`binary: bootstrap`, `goos: [linux]`, `goarch: [arm64]`, `flags: [-trimpath]`).
     - Archive it as a zip with `name_template: lambda` (to get `lambda.zip`), `wrap_in_directory: false` and `files: [LICENSE.md]`, with `bootstrap` at the zip root as Lambda requires.
     - The `upload` and `files:` inputs of the old path go away. The Terraform module downloads `.../releases/download/<version>/lambda.zip` from the GitHub release, so the release asset must stay named exactly `lambda.zip`. It also checks a local `dist/lambda.zip` (`terraform/module/locals.tf`), so keep `task build` producing that path for local use, or point it at GoReleaser's `dist/` output.

**Keeping "build everything, upload only what we ship" (the commented-out requirement)**
- Define two build ids in `.goreleaser.yaml`: `release` with `darwin/arm64` only, and `matrix` with the other five targets.
- Set `archives.ids: [release]` and `checksum.ids: [release]` (with the binary format) so only the `release` build is uploaded. The `matrix` builds still compile in every run, which shows the code builds on those targets.
- To start shipping the other targets later, add `matrix` to `archives.ids` (and `checksum.ids`). It is a one-line change per repo. In `.goreleaser.yaml`, include that line commented out, with a comment explaining it:

  ```yaml
  archives:
    - formats: [binary]
      ids:
        - release
        # Uncomment to publish the other platforms (macOS Intel, Linux, Windows) in releases.
        # They are built on every run to prove they compile, but not uploaded yet.
        # - matrix

  checksum:
    ids:
      - release
      # Uncomment together with `matrix` under `archives.ids` above.
      # - matrix
  ```

  Check the exact key names (`formats` versus the older `format`) against the GoReleaser version in use.
- Verify `ids` filtering behaves this way with a snapshot build, since I haven't run GoReleaser here.

**Cost and risk**: low to moderate. The config is about 30 lines. The main work is the shared workflow and the tag handling, and doing it in `go-git-helper` first as the pilot would expose the issues before the other two repos move.

### Phase 2: Make `update` cross-platform

1. Replace the hardcoded path with the running binary's location, using `os.Executable()` followed by `filepath.EvalSymlinks`. Keep an override flag such as `--install-path`.
2. Build the asset name from `runtime.GOOS` and `runtime.GOARCH`, and add `.exe` on Windows. Merge the duplicated `darwin` and `linux` cases. Return an error for any combination with no matching release asset, and stop instead of continuing with an empty URL.
3. Download to `os.CreateTemp` in the target directory. Check the HTTP status, optionally verify the checksum, and `chmod 0755` on Unix.
4. Replace the binary:
   - Unix: `os.Rename` of the temp file over the target. Fall back to `sudo mv` only if the directory is not writable (with a clear message). Never use `sudo` on Windows.
   - Windows: rename the running `.exe` to `git-helper.exe.old`, move the new file into place, and clean up the `.old` file on the next run.
5. Drop `chown user:staff` and `sudo chmod`. Verify the new version by running the new binary by absolute path.
6. Update `update_test.go` for the new flow. Use an `httptest` server and a temp dir rather than asserting `sudo` calls.

### Phase 3: Make `setup` cross-platform

1. Plugins:
   - Quote `"$@"` in all `plugins/git-*` scripts.
   - Unix and Git Bash: keep the shell scripts.
   - Windows: also generate `git-<name>.cmd` shims (`@git-helper <name> %*`) so the commands work from `cmd` and PowerShell.
   - Embed the plugins with `go:embed` instead of fetching them from the GitHub contents API (decided). This removes the network dependency and allows generating per-OS shims.
2. Print per-OS next steps: `PATH` instructions for bash, zsh and fish, and for PowerShell (`$env:Path` or the system settings dialog). Do the same for completion activation in each shell.
3. Use `filepath.Join` for every path. Write the config with `0600` and the dir with `0700`.
4. Offer the `~/.git-helper/plugins` `PATH` change as printed instructions only. Do not edit shell profiles automatically.

### Phase 4: Robustness fixes found along the way

- Fix the nil `info` dereference in `ConfigDirExists`.
- Make `git` output parsing tolerant of `\r`.
- Remove the dead `os.Stdout` reassignment in `waitAndStdout`.
- Give a clear error if `git` is not found on `PATH` (`exec.LookPath`).

### Phase 5: Tests

- Isolate home dir in tests with `t.Setenv("HOME", d)` and `t.Setenv("USERPROFILE", d)`. Stop reading the real home directory (`gitlab_test.go`, `codeRequest_test.go`).
- Use `filepath.Join` in assertions.
- Add table tests for asset-name and install-path selection across all GOOS/GOARCH pairs (parameterised, so the Windows case runs on any host).
- Run the full suite on the Linux, macOS (arm64 and Intel) and Windows CI jobs.

### Phase 6: Docs

Rewrite the README Installation section:

- Per-OS download table (asset names above).
- macOS (Apple Silicon and Intel): Gatekeeper notes. Consider code-signing and notarization later, or `xattr -d com.apple.quarantine`.
- Linux: `install -m 0755 git-helper_linux_amd64 ~/.local/bin/git-helper` (no root required) or `/usr/local/bin` with `sudo`.
- Windows: put `git-helper.exe` on `PATH` (for example `%USERPROFILE%\bin`), SmartScreen notes, and Git for Windows as a prerequisite.
- Shell aliases and completions for zsh, bash, fish and PowerShell.
- Remove the "Apple Silicon only, no guarantees" notice once the CI matrix is green.

## 4. Suggested Order and Risk

| Order | Item | Effort | Risk |
|---|---|---|---|
| 1 | Phase 1 (build, release, CI matrix) | S–M | Low. It may require a change in the shared workflow repo. |
| 2 | Phase 4 and 5 (bug fixes, test isolation) | S | Low |
| 3 | Phase 2 (`update`) | M | Medium. Windows self-replace needs real-hardware testing. |
| 4 | Phase 3 (`setup`) | M | Medium. Needs a decision on the Windows plugin approach. |
| 5 | Phase 6 (docs) | S | Low |

macOS Intel and Linux are mostly covered by Phases 1, 2 and 6. Windows is the bulk of the remaining work, in Phases 2 and 3.

## 5. Decisions

1. **Plugins: embed for all systems.** Use `go:embed` for the plugin scripts so `setup` needs no network call. On Windows, also generate `.cmd` shims from the embedded list. Phase 3 is updated to match.
2. **Config location: keep `~/.git-helper` on all platforms (default, pending your confirmation).**
   - Platforms have different conventions. macOS tools commonly use `~/Library/Application Support` or `~/.config`. Linux uses `$XDG_CONFIG_HOME`, which is usually `~/.config`. Windows uses `%AppData%`. Go's `os.UserConfigDir()` follows these.
   - You said you want "wherever the machine keeps their core dot config files". On all three systems that is the home directory: `~/.gitconfig`, `~/.ssh` and so on, including Windows via Git for Windows. `~/.git-helper` sits in that same place, and existing users keep their config.
   - If you would rather follow each OS convention, use `os.UserConfigDir()` with a fallback to the legacy `~/.git-helper`. That adds a migration path and more tests.
3. **`update` and `sudo`.**
   - Moving into `/usr/local/bin` needs `sudo` on Linux too, because it is root-owned there as well. Windows has no `sudo` and does not need it for a user-owned directory.
   - Update the binary where it currently lives (`os.Executable()`) and try a plain rename first. If that fails with a permission error on macOS or Linux, fall back to `sudo mv`, with a message that says why. Windows never uses `sudo`.
   - The macOS behaviour you rely on is unchanged: a binary in `/usr/local/bin` still prompts for `sudo`. Users with a user-owned install directory no longer need it.
4. **Architectures: include `linux/arm64` and `windows/arm64`.** It is cheap to build them since they are pure-Go cross-compiles. Phase 1 builds all six targets as a smoke test. Test runners cover macOS arm64 and Intel, Linux amd64 and Windows amd64. Add `ubuntu-24.04-arm` and `windows-11-arm` runners only if the shared workflow supports a matrix. Otherwise those two are build-only.
5. **No Homebrew tap or Scoop bucket.** The optional follow-up is removed.
6. **No signing or notarization.** The README keeps the Gatekeeper and SmartScreen "allow anyway" notes.
