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
Usage of ./prow-aliases-syncer:
  -b, --branch strings   branch to update (glob expression supported) (can be given multiple times)
      --dry-run          do not actually push to GitHub (repositories will still be cloned and locally updated)
  -k, --keep             keep unknown teams (do not combine with -strict)
  -o, --org string       GitHub organization to work with
  -s, --strict           compare owners files byte by byte
  -u, --update           do not create pull requests, but directly push into the target branches
  -v, --verbose          Enable more verbose output
```

For example:

```bash
$ export GITHUB_TOKEN=ghp_....
$ prow-aliases-syncer --org myorg --strict --branch master --branch main --branch 'release/*'
```

## License

MIT
