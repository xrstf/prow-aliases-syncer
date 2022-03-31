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
  -o, --org string       GitHub organization to work with
  -s, --strict           compare owners files byte by byte
  -v, --verbose          Enable more verbose output
```

## License

MIT
