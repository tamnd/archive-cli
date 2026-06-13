---
title: "archive"
description: "A fast, friendly command line for the Internet Archive. Search millions of items, read metadata, download and verify files, upload to your own items, read view counts, and travel through the Wayback Machine, all from one binary."
heroTitle: "The Internet Archive, from the command line"
heroLead: "archive is a single pure-Go binary that puts archive.org behind a tool that feels like curl. Search the whole collection, inspect an item, download files with checksum verification, push your own uploads, and pull any page out of the Wayback Machine, with no credentials needed for the public data."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

Working with the Internet Archive usually means juggling the Metadata API, the
Solr search endpoint, S3-style upload headers, and the Wayback CDX server by
hand. archive puts all of it behind one tool with sensible defaults, real
output formats, and pipelines that compose.

```bash
archive search 'collection:nasa' -n 5      # find items
archive item nasa                          # what an item is, at a glance
archive files nasa --format JPEG -o url     # a file listing, as plain URLs
archive download nasa --format JPEG -d .    # download and verify by md5
archive wayback get example.com -t 2010     # a page as it was in 2010
```

It talks to the public data on `archive.org` over HTTPS, so there is nothing to
sign up for. The binary is pure Go with no runtime dependencies. Credentials
are only needed to upload, delete, or read your task queue; everything else is
anonymous.

## What you can do with it

- **Search.** Query the Advanced Search (Solr) index for any Lucene query,
  sort and project fields, and render the result as a table, JSONL, CSV, or
  just identifiers. Large result sets page automatically through the Scraping
  API.
- **Inspect items.** Read the raw Metadata API document, a friendly summary, or
  a single field, and list the files in an item filtered by glob or format.
- **Download and verify.** Pull whole items or selected files concurrently,
  resume partial downloads with HTTP range requests, and verify each file
  against its md5.
- **Upload and manage.** Push files into your own items over the S3-like IAS3
  interface with metadata headers, and delete files when you need to.
- **Travel the Wayback Machine.** Find the closest snapshot of a URL, list its
  full capture history from the CDX server, fetch a snapshot as text, links, or
  raw bytes, and trigger a fresh capture with Save Page Now.

## Where to go next

- New here? Start with the [introduction](/getting-started/introduction/) for
  the mental model, then the [quick start](/getting-started/quick-start/).
- Want to install it? See [installation](/getting-started/installation/).
- Looking for a specific task? The [guides](/guides/) cover searching, items
  and metadata, downloading, uploading, and the Wayback Machine.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
