---
title: "The Wayback Machine"
description: "Find the closest snapshot of a URL, list its capture history, fetch a snapshot as text or links, and save a fresh capture."
weight: 50
---

The `wayback` group (alias `wb`) works with the Wayback Machine: the web-capture
side of the Internet Archive, addressed by URL and timestamp rather than by item
identifier.

## The closest snapshot

```bash
archive wayback available example.com
archive wayback available example.com -t 2010
```

`available` asks the Availability API for the capture nearest a timestamp
(`-t`, a full or partial `YYYYMMDDhhmmss`). It returns the snapshot's timestamp,
HTTP status, and replay URL.

## The full capture history

```bash
archive wayback list example.com -n 10
archive wayback list example.com --from 2010 --to 2012 --status 200
archive wayback cdx 'example.com/*' --match-type prefix --collapse digest
```

`list` (alias `cdx`) reads the CDX server, which returns one row per capture:
timestamp, original URL, MIME type, status, digest, and length. Narrow it with
`--from`/`--to`, `--status`, `--mime`, a raw `--filter`, or `--collapse` to fold
adjacent duplicate rows on a field.

The CDX server is aggressively rate-limited by the Archive. archive throttles
and retries with backoff automatically, but a busy moment can still return
429s; raise `--rate` and `--retries` and try again.

## Fetching a snapshot

```bash
archive wayback get example.com -t 2010 --text     # readable text
archive wayback get example.com -t 2010 --links     # the page's hyperlinks
archive wayback get example.com -t 2010 --raw > page.html  # original bytes
archive wayback get example.com -t 2010 -o page.html        # write to a file
```

`get` resolves the closest snapshot (or the one at `-t`), then fetches the
original archived bytes. `--text` extracts readable text, `--links` lists the
hyperlinks (great with `-o url`), and the default is the raw archived HTML.

## Saving a fresh capture

```bash
archive wayback save https://example.com/
```

Anonymously this is a fire-and-forget request to Save Page Now. With
`--outlinks` or `--screenshot` it uses the authenticated SPN2 API (which needs
[credentials](/reference/configuration/)) and, with `--wait`, polls the capture
job to completion:

```bash
archive wayback save https://example.com/ --outlinks --wait
```
