# git-reviewers

Figure out who would be a good reviewer for a change.

There are two places checked when determining reviewers:

- The results of `git blame` and the authors mentioned in the changes.
- The person that merged changes last in the file

## Usage

```shell
Usage:
  git-reviewers [flags]

Flags:
  -h, --help      help for git-reviewers
  -v, --verbose   Increase verbosity
      --version   version for git-reviewers
```
