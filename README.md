# archive

A fast, friendly command line for the [Internet Archive](https://archive.org).
One binary that searches millions of items, reads metadata, downloads and
verifies files, uploads to your own items, reads view counts, and travels
through the Wayback Machine.

```
archive item nasa
```

```
field       value
identifier  nasa
title       NASA
mediatype   collection
files       9
size        131.1 KB
server      ia801607.us.archive.org
details     https://archive.org/details/nasa
```

Full documentation: [archive-cli.tamnd.com](https://archive-cli.tamnd.com).

## Why

Working with the Internet Archive usually means juggling the Metadata API, the
Solr search endpoint, S3-style upload headers, and the Wayback CDX server by
hand. archive puts all of it behind one tool with sensible defaults, real output
formats, and pipelines that compose. It talks to the public data on
`archive.org` over HTTPS, so there is nothing to sign up for; credentials are
only needed to upload, delete, or read your task queue.

## Install

```sh
go install github.com/tamnd/archive-cli/cmd/archive@latest
```

Or grab a prebuilt binary, a Linux package (`deb`/`rpm`/`apk`), or a container
image from the [releases page](https://github.com/tamnd/archive-cli/releases).
The binary is pure Go with no runtime dependencies.

```sh
brew install tamnd/tap/archive                 # macOS / Linux
docker run --rm ghcr.io/tamnd/archive search nasa -n 5
```

Build from source:

```sh
git clone https://github.com/tamnd/archive-cli
cd archive-cli
make build      # produces ./bin/archive
```

## Quick start

```sh
archive search 'collection:nasa' -n 5      # find items
archive item nasa                          # what an item is, at a glance
archive metadata nasa metadata/title        # a single metadata field
archive files nasa --format JPEG -o url     # a file listing, as URLs
archive download nasa --format JPEG -d .     # download and verify by md5
archive views nasa                          # view counts
archive wayback get example.com -t 2010 --text  # a page as it was in 2010
```

## What you can do with it

- **Search.** Query the Advanced Search (Solr) index with any Lucene query,
  sort and project fields, and render the result as a table, JSONL, CSV, or just
  identifiers. Large sets export through the cursor-based Scraping API with
  `--all`.
- **Inspect items.** Read the raw Metadata API document, a friendly summary, or
  a single field, and list files filtered by glob or format.
- **Download and verify.** Pull whole items or selected files concurrently,
  resume partial downloads with HTTP range requests, and verify each file
  against its md5.
- **Upload and manage.** Push files into your own items over the S3-like IAS3
  interface with metadata headers, and delete files.
- **Travel the Wayback Machine.** Find the closest snapshot of a URL, list its
  capture history from the CDX server, fetch a snapshot as text, links, or raw
  bytes, and trigger a fresh capture with Save Page Now.

## Output formats

Every command renders through one output layer. Pick with `-o`: `table`, `json`,
`jsonl`, `csv`, `tsv`, `url`, or `raw`. `auto` (the default) is a table on a
terminal and JSONL in a pipe. `--fields` projects columns; `--template` applies
a Go template per row.

```sh
archive search 'collection:nasa' --fields identifier,downloads -o csv
archive search 'collection:nasa' --fields identifier -o raw | xargs -n1 archive item
```

## Credentials

Reading public data needs no account. To upload, delete, or read a private task
queue, get an IAS3 key pair from
[archive.org/account/s3.php](https://archive.org/account/s3.php) and store it:

```sh
archive configure                 # prompts, writes ~/.config/archive/credentials (0600)
archive whoami                    # show what is configured
```

Credentials resolve from `--access`/`--secret`, then `ARCHIVE_ACCESS_KEY` /
`ARCHIVE_SECRET_KEY` (or `IA_*`), then the credentials file.

## Exit codes

`0` success, `1` generic error, `2` usage error, `3` no results, `4`
authentication required/failed, `5` not found.

## Development

```sh
make build      # build ./bin/archive
make test       # go test ./...
make vet        # go vet ./...
make fmt        # gofmt -w -s .
```

CI runs build, test (with the race detector) on Linux and macOS, gofmt, vet,
golangci-lint, govulncheck, and a go.mod tidiness check. Releases are cut by
pushing a `vX.Y.Z` tag, which GoReleaser turns into archives, Linux packages, a
multi-arch GHCR image, checksums, an SBOM, a cosign signature, and Homebrew and
Scoop entries.

## License

[Apache-2.0](LICENSE).
