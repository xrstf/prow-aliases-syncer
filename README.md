# prow-aliases-syncer

This repository contains a Go program to synchronize GitHub teams to the
`OWNERS_ALIASES` file used by Prow.

## Installation

You can download a binary for the [latest release on GitHub](https://github.com/xrstf/prow-aliases-syncer/releases)
or install it via Go:

```bash
go install go.xrstf.de/prow-aliases-syncer
```

## Usage

```
Usage of _build/prow-aliases-syncer:
      --body string           file with a template for the PR body
  -b, --branch strings        branch to update (glob expression supported) (can be given multiple times)
      --dry-run               do not actually push to GitHub (repositories will still be cloned and locally updated)
      --header string         file with header for the generated aliases files
  -i, --ignore-user strings   GitHub usernames which should be ignored when determining the most recent commit on branch (can be given multiple times)
  -k, --keep                  keep unknown teams (do not combine with -strict)
      --max-age duration      only update branches with commits within this duration (default 2160h0m0s)
  -o, --org string            GitHub organization to load teams from and update repositories in (unless --target-org is given)
  -s, --strict                compare owners files byte by byte
  -t, --target-org string     update repositories in this org based on the teams from --org
  -u, --update                do not create pull requests, but directly push into the target branches
  -v, --verbose               Enable more verbose output
  -V, --version               show version info and exit immediately
```

For example:

```bash
$ export GITHUB_TOKEN=ghp_....
$ prow-aliases-syncer --org myorg --strict --branch main --branch 'release/*'
```

## License

MIT
