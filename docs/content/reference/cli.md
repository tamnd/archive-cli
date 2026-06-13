---
title: "CLI reference"
description: "Every archive command, its arguments, and its flags."
weight: 10
---

This is the full command surface. Every command also accepts the
[global flags](#global-flags) and prints its own `--help`.

## Global flags

These persistent flags apply to every command:

| Flag | Default | Meaning |
|------|---------|---------|
| `-o, --output` | `auto` | `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw` |
| `--fields` | | comma-separated columns to show |
| `--template` | | Go `text/template` applied per row |
| `-n, --limit` | `0` | max results (0 = unlimited) |
| `--no-header` | `false` | omit the header row in table/csv output |
| `--data-dir` | `~/data/archive` | root data directory |
| `-j, --workers` | `8` | concurrency |
| `--rate` | `250ms` | minimum delay between requests |
| `--retries` | `5` | retry attempts on 429/5xx |
| `--timeout` | `2m` | per-request timeout |
| `--no-cache` | `false` | bypass on-disk caches |
| `--color` | `auto` | `auto`, `always`, or `never` |
| `-q, --quiet` | `false` | suppress progress output |
| `-v, --verbose` | | increase verbosity (repeatable) |
| `-y, --yes` | `false` | assume yes to prompts |
| `--dry-run` | `false` | print actions without performing them |
| `--access`, `--secret` | | IAS3 credentials (override config/env) |

## search

`archive search <query> [flags]` — search the items index (Advanced Search /
Solr). The query is Lucene.

| Flag | Meaning |
|------|---------|
| `--media` | filter by mediatype |
| `--collection` | filter by collection |
| `--creator` | filter by creator |
| `--year` | filter by year or range `A-B` |
| `--sort` | sort key, e.g. `downloads desc` (repeatable) |
| `-f, --field` | metadata field to return (repeatable) |
| `--count` | print only the number of matches |
| `--all` | export every match via the cursor (Scraping API) |
| `--rows` | page size for Advanced Search |

## item

`archive item <identifier>` — a friendly one-table summary of an item.

## metadata

`archive metadata <identifier> [subpath]` — the raw Metadata API document, or a
single slash-path subpath such as `metadata/title` or `files`.

## files

`archive files <identifier> [flags]` — list the files in an item.

| Flag | Meaning |
|------|---------|
| `--glob` | filter file names by glob, e.g. `'*.pdf'` |
| `--format` | filter by format substring, e.g. `PDF` |

## download

`archive download <identifier|-> [files...] [flags]` — download files from an
item into `<out-dir>/<identifier>/`. With `-` as the identifier, identifiers are
read from stdin.

| Flag | Meaning |
|------|---------|
| `-d, --out-dir` | destination directory (default `~/data/archive/download`) |
| `--glob` | only files whose name matches this glob |
| `--format` | only files whose format contains this substring |
| `--verify` | verify each download against its md5 |
| `--flat` | drop sub-directories from file names |

## upload

`archive upload <identifier> <file...> [flags]` — upload files over IAS3.
Requires credentials.

| Flag | Meaning |
|------|---------|
| `-m, --metadata` | metadata `key:value` to set on the item (repeatable) |
| `--make-bucket` | create the item if it does not exist |
| `--name` | remote file name (single-file uploads only) |
| `--content-type` | override the Content-Type header |
| `--no-derive` | skip derivation after upload |

## delete

`archive delete <identifier> <remote-file...>` — delete files from an item over
IAS3. Requires credentials. Prompts unless `-y`.

## wayback

`archive wayback <subcommand>` (alias `wb`) — work with the Wayback Machine.

- `wayback available <url> [-t timestamp]` — the closest archived snapshot.
- `wayback list <url> [flags]` (alias `cdx`) — capture history from the CDX
  server. Flags: `--from`, `--to`, `--match-type`, `--filter` (repeatable),
  `--collapse`, `--status`, `--mime`.
- `wayback get <url> [flags]` — fetch a snapshot. Flags: `-t/--timestamp`,
  `--raw` (default), `--text`, `--links`, `--out`.
- `wayback save <url> [flags]` — Save Page Now. Flags: `--outlinks`,
  `--screenshot` (both SPN2, need auth), `--wait`.

## views

`archive views <identifier...>` — view statistics (all-time / 30-day / 7-day)
for one or more items.

## tasks

`archive tasks <identifier>` — the catalog/derive task history of an item. Needs
credentials for items you do not own.

## open

`archive open <identifier|url> [--web]` — print the details or Wayback URL for a
thing; `--web` opens it in the default browser.

## configure

`archive configure [--access ... --secret ...]` — store IAS3 credentials in
`~/.config/archive/credentials` (mode 0600). Prompts when keys are not passed.

## whoami

`archive whoami` — show the configured credentials (secret masked) and where
they came from.

## config

`archive config show` — print the effective configuration and the data paths a
run will use.

## cache

`archive cache <subcommand>` — `dir` prints the cache directory; `info` shows
its size; `clear` empties it.

## version

`archive version [--short]` — print version, commit, build date, and toolchain.
