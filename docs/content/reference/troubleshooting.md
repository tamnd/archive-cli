---
title: "Troubleshooting"
description: "Rate limits, exit codes, and the errors you are most likely to hit."
weight: 40
---

## Exit codes

archive uses distinct exit codes so scripts can branch on the kind of failure:

| Code | Meaning |
|------|---------|
| `0` | success |
| `1` | generic error (network, unexpected response) |
| `2` | usage error (unknown command or flag, bad arguments) |
| `3` | no results (an empty search, no captures, nothing matched) |
| `4` | authentication required or failed |
| `5` | not found (no such item, file, or snapshot) |

```bash
archive item maybe-missing || echo "exit $?"
```

## "Too Many Requests" (HTTP 429)

The Wayback **CDX** and **replay** endpoints are rate-limited hard by the
Internet Archive, and the limit is per-IP, so a shared or datacenter address can
be throttled even on your first request. archive already throttles requests and
retries with backoff that honours the server's `Retry-After`, but a busy moment
can still exhaust the retries.

If you see persistent 429s:

- Slow down with `--rate 2s` (or higher) and raise `--retries`.
- Wait a minute or two for the per-IP limit to reset, then try again.
- For the items API (`search`, `metadata`, `files`, `views`), 429s are rare;
  the default rate is plenty.

## "authentication required" (exit 4)

`upload`, `delete`, `wayback save --outlinks/--screenshot`, and `tasks` on items
you do not own need IAS3 credentials. Run `archive configure`, or set
`ARCHIVE_ACCESS_KEY` / `ARCHIVE_SECRET_KEY`. Check with `archive whoami`. See
[configuration](/reference/configuration/).

## "not found" (exit 5)

The identifier, file name, or snapshot does not exist. Double-check the
identifier with `archive search`, and remember identifiers are case-sensitive.

## Stale data

Metadata and search responses are cached briefly on disk. If you just changed
an item and see the old state, bypass the cache for one run with `--no-cache`,
or clear it entirely:

```bash
archive --no-cache item my-item
archive cache clear
```

## Seeing what a command will do

`--dry-run` prints the actions (and, for `upload`, the exact request and
headers) without performing them. `-v`/`--verbose` raises the detail of progress
output. `archive config show` prints every resolved path and setting.
