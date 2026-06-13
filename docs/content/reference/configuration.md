---
title: "Configuration"
description: "Data paths, environment variables, and how credentials resolve."
weight: 20
---

archive needs no configuration to read public data. Everything below is
optional: where it keeps state, and how it finds credentials when you upload,
delete, or read a task queue.

## Where state lives

```bash
archive config show
```

| Path | Default | Override |
|------|---------|----------|
| Data root | `~/data/archive` | `--data-dir`, `ARCHIVE_DATA_DIR` |
| Downloads | `<data>/download` | `download -d` |
| Cache | `<data>/cache` | follows the data root |
| Config | `~/.config/archive` | `XDG_CONFIG_HOME` |
| Credentials | `~/.config/archive/credentials` | |

`config show` prints the resolved values plus the effective client settings
(workers, rate, timeout, retries) and whether credentials are loaded.

## Credentials

Authenticated commands (`upload`, `delete`, `wayback save --outlinks/--screenshot`,
and `tasks` on items you do not own) use an IAS3 access/secret key pair from
[archive.org/account/s3.php](https://archive.org/account/s3.php).

Store them once:

```bash
archive configure --access YOUR_KEY --secret YOUR_SECRET
```

With no flags, `configure` prompts (the secret is read without echo). You can
also log in with your account email and password to fetch the keys:

```bash
archive configure --email you@example.com
```

Credentials are written to `~/.config/archive/credentials` with `0600`
permissions. Check what is configured (the secret is masked):

```bash
archive whoami
```

### Resolution order

For any command that needs credentials, archive resolves them in this order,
first match wins:

1. `--access` / `--secret` flags.
2. `ARCHIVE_ACCESS_KEY` / `ARCHIVE_SECRET_KEY` environment variables.
3. `IA_ACCESS_KEY` / `IA_SECRET_KEY` (the names the `ia` Python tool uses).
4. The `~/.config/archive/credentials` file.

This lets you keep a stored default and override it per command, or run fully
from the environment in CI without writing a file.

## Networking knobs

| Flag | Default | What it does |
|------|---------|--------------|
| `--rate` | `250ms` | minimum delay between requests |
| `--retries` | `5` | backoff retries on 429/5xx (honours `Retry-After`) |
| `--timeout` | `2m` | per-request timeout |
| `-j, --workers` | `8` | concurrent downloads |

The Wayback CDX and replay endpoints are rate-limited hard by the Archive; if
you hit 429s, raise `--rate` (e.g. `--rate 2s`) and `--retries`.

## Caching

Metadata, search pages, and availability lookups are cached on disk under the
cache directory, keyed by request with a short TTL, so repeated commands are
instant and gentle on the Archive. Bypass it for one run with `--no-cache`, and
manage it with the `cache` command:

```bash
archive cache info     # size and entry count
archive cache clear     # empty it
```
