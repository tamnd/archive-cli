# archive

[![CI](https://github.com/tamnd/archive-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/archive-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tamnd/archive-cli)](https://github.com/tamnd/archive-cli/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tamnd/archive-cli.svg)](https://pkg.go.dev/github.com/tamnd/archive-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamnd/archive-cli)](https://goreportcard.com/report/github.com/tamnd/archive-cli)
[![License](https://img.shields.io/github/license/tamnd/archive-cli)](./LICENSE)

A command line for the [Internet Archive](https://archive.org). `archive`
searches millions of items, reads metadata, downloads and verifies files,
travels through the Wayback Machine, and uploads to your own items. One
pure-Go binary, no credentials needed for public data.

[Install](#install) • [Commands](#commands) • [Usage](#usage) • [Credentials](#credentials)

![archive searching the Internet Archive and querying the Wayback Machine](docs/static/demo.gif)

It talks to the public Internet Archive APIs over HTTPS: the Metadata API, the
Advanced Search (Solr) endpoint, the CDX Wayback server, and the S3-like IAS3
upload interface. Every request is paced, retried on transient failures, and
cached on disk. No login is needed for anything read-only.

`archive` is an independent tool. It is not affiliated with or endorsed by the
Internet Archive.

## Install

```bash
go install github.com/tamnd/archive-cli/cmd/archive@latest
```

Or grab a prebuilt binary, a Linux package (`deb`/`rpm`/`apk`), or a container
image from the [releases](https://github.com/tamnd/archive-cli/releases):

```bash
brew install tamnd/tap/archive
docker run --rm ghcr.io/tamnd/archive:latest search 'collection:nasa' -n 5
```

Shell completion is built in: `archive completion bash|zsh|fish|powershell`.

## Commands

| Command | Does |
| --- | --- |
| `archive search <query>` | search the Solr index; any Lucene query, `--all` for cursor-based export |
| `archive item <identifier>` | a friendly summary of an item |
| `archive metadata <identifier> [subpath]` | the raw Metadata API document, or one field |
| `archive files <identifier>` | files in an item; `--format`, `--glob` to filter |
| `archive download <identifier> [files...]` | download and md5-verify; `--workers`, `-d` dir |
| `archive upload <identifier> <file...>` | upload into an item over IAS3; `--meta` |
| `archive delete <identifier> <file...>` | delete files from an item over IAS3 |
| `archive views <identifier...>` | view statistics for one or more items |
| `archive tasks <identifier>` | catalog and derive task history of an item |
| `archive wayback available <url>` | closest archived snapshot of a URL |
| `archive wayback list <url>` | capture history from the CDX server |
| `archive wayback get <url>` | fetch the content of a snapshot; `--text`, `-t` timestamp |
| `archive wayback save <url>` | trigger a fresh Save Page Now capture |
| `archive open <identifier\|url>` | open the details or Wayback URL in the browser |
| `archive configure` | store IAS3 credentials |
| `archive whoami` | show configured credentials |
| `archive config` | show resolved configuration and data paths |
| `archive cache path\|info\|clear` | inspect or clear the on-disk cache |
| `archive version` | print version information |

Full reference and guides live at [archive-cli.tamnd.com](https://archive-cli.tamnd.com).

## Usage

```bash
archive search 'collection:nasa' -n 10              # find items
archive item nasa                                   # item summary
archive metadata nasa metadata/title                # one metadata field
archive files nasa --format JPEG -o url             # file listing as URLs
archive download nasa --format JPEG -d .            # download and verify
archive views nasa                                  # view statistics
archive wayback get https://example.com -t 2010     # a page as it was in 2010
```

Records come out as a table (the default on a terminal), JSON, JSONL, CSV, TSV,
url, or raw:

```bash
archive search 'collection:nasa' --fields identifier,title,downloads -o table
archive search 'collection:nasa' --fields identifier,downloads -o csv
archive search 'collection:nasa' --fields identifier -o raw | xargs -n1 archive item
archive files nasa --format JPEG -o url | head -20
archive wayback list https://archive.org -n 50 -o jsonl | jq .timestamp
```

Export a large result set with the cursor-based Scraping API:

```bash
archive search 'subject:jazz mediatype:audio' --all --fields identifier,title -o jsonl > jazz.jsonl
```

### Global flags

```
-o, --output    table|json|jsonl|csv|tsv|url|raw   (auto: table on a TTY, jsonl when piped)
    --fields    comma-separated columns to include
    --no-header omit the header row in table/csv/tsv
    --template  Go text/template applied per record
-n, --limit     max records (0 = unlimited)
-q, --quiet     suppress progress output
    --color     auto|always|never
    --rate      min spacing between requests (default 250ms)
    --timeout   per-request timeout (default 2m)
    --retries   retry attempts on 429/5xx (default 5)
-j, --workers   concurrency for downloads (default 8)
    --no-cache  bypass the on-disk cache
    --dry-run   print actions without performing them
```

## Credentials

Reading public data needs no account. To upload, delete, or read a private task
queue, get an IAS3 key pair from
[archive.org/account/s3.php](https://archive.org/account/s3.php) and store it:

```bash
archive configure    # prompts for access and secret keys, writes ~/.config/archive/credentials
archive whoami       # verify what is configured
```

Credentials resolve in order: `--access`/`--secret` flags, then
`ARCHIVE_ACCESS_KEY`/`ARCHIVE_SECRET_KEY` environment variables (or `IA_*`
aliases), then the credentials file.

## Exit codes

```
0  success
1  error
2  usage error
3  no results
4  authentication required or failed
5  not found
```

## Development

```
cmd/archive/    thin main entry point
cli/            cobra commands and output rendering
ia/             HTTP client, API calls, and models
docs/           documentation site (Hugo, tago-doks theme)
```

```bash
make build   # ./bin/archive
make test    # go test ./...
make vet     # go vet ./...
make fmt     # gofmt -w -s .
```

Requires Go 1.23+.

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds archives,
Linux packages, a multi-arch GHCR image, checksums, an SBOM, a cosign
signature, and Homebrew and Scoop entries:

```bash
git tag -a v0.2.0 -m "v0.2.0"
git push --tags
```

The image tag carries no `v` prefix (`ghcr.io/tamnd/archive:0.2.0`).

## License

Apache-2.0. See [LICENSE](LICENSE).
