# Contributing

To contribute to this project please carefully read this document.

## Setup

`wakatime-cli` is written in [Go](https://golang.org/).

Prerequisites:

- We use `make` to build and run tests
- We use `bats` to test shell scripts. Documentation can be found [here](https://bats-core.readthedocs.io/en/latest/installation.html)
- [Go 1.16](https://golang.org/doc/install)

After cloning, install dependencies with `make install`.

## Branches

This project currently has two branches

- `develop` - Default branch for every new `feature` or `fix`
- `release` - Branch for production releases and hotfixes

## Testing and Linting

Run `make test-all` before creating any pull requests, or your PR won’t pass the automated checks.

> make sure you build binary by setting its version otherwise integration tests will fail. `VERSION=v0.0.1-test make build-<os>-<architecture>`. For testing shell scripts you might initialize submodules by running `git submodule update --init --recursive`.

## Branching Stratgegy

Please follow our guideline for branch names [here](https://github.com/wakatime/semver-action#branch-names). Branches off the pattern won't be accepted.

## Pull Requests

- Big changes, changes to the API, or changes with backward compatibility trade-offs should be first discussed in the Slack.
- Search [existing pull requests](https://github.com/wakatime/wakatime-cli/pulls) to see if one has already been submitted for this change. Search the [issues](https://github.com/wakatime/wakatime-cli/issues?q=is%3Aissue) to see if there has been a discussion on this topic and whether your pull request can close any issues.
- Code formatting should be consistent with the style used in the existing code.
- Don't leave commented out code. A record of this code is already preserved in the commit history.
- All commits must be atomic. This means that the commit completely accomplishes a single task. Each commit should result in fully functional code. Multiple tasks should not be combined in a single commit, but a single task should not be split over multiple commits (e.g. one commit per file modified is not a good practice). For more information see <http://www.freshconsulting.com/atomic-commits>.
- Each pull request should address a single bug fix or feature. This may consist of multiple commits. If you have multiple, unrelated fixes or enhancements to contribute, submit them as separate pull requests.
- Commit messages:
  - Use the [imperative mood](http://chris.beams.io/posts/git-commit/#imperative) in the title. For example: "Apply editor.indent preference"
  - Capitalize the title.
  - Do not end the title with a period.
  - Separate title from the body with a blank line. If you're committing via GitHub or GitHub Desktop this will be done automatically.
  - Wrap body at 72 characters.
  - Completely explain the purpose of the commit. Include a rationale for the change, any caveats, side-effects, etc.
  - If your pull request fixes an issue in the issue tracker, use the [closes/fixes/resolves syntax](https://help.github.com/articles/closing-issues-via-commit-messages) in the body to indicate this.
  - See <http://chris.beams.io/posts/git-commit> for more tips on writing good commit messages.
- Pull request title and description should follow the same guidelines as commit messages.
- Rebasing pull requests is OK and encouraged. After submitting your pull request some changes may be requested. Prefer using [git fixup](https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---fixupltcommitgt) rather than adding orphan extra commits to the pull request, then do a push to your fork. As soon as your PR gets approved one of us will merge it by rebasing and squashing any residuary commits that were pushed while reviewing. This will help to keep the commit history of the repository clean.

Any question join us on [Slack](https://wakaslack.herokuapp.com/).
