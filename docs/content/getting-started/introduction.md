---
title: "Introduction"
description: "What the Internet Archive is, and the handful of ideas archive is built around."
weight: 10
---

The [Internet Archive](https://archive.org) is a non-profit digital library:
millions of books, movies, audio recordings, software, and web pages, all
freely accessible. archive is a command-line client for it. This page is the
mental model; the [quick start](/getting-started/quick-start/) is the hands-on
version.

## Items, identifiers, and files

Everything on archive.org is an **item**, and every item has a unique
**identifier**: a short slug like `nasa` or `goody`. An item is really a
directory of **files** plus a **metadata** record describing it.

```bash
archive item nasa        # a friendly summary of the item
archive files nasa       # the files inside it
archive metadata nasa    # the raw metadata document
```

The identifier is the one thing you need to address an item. Most commands take
it as their first argument.

## Mediatypes and collections

Each item has a **mediatype** (`texts`, `movies`, `audio`, `image`, `software`,
`web`, `data`, or `collection`) and belongs to one or more **collections**.
Collections are themselves items with `mediatype:collection`, which is why you
can search inside one with a Lucene query:

```bash
archive search 'collection:nasa AND mediatype:image' -n 10
```

## Search is Lucene

archive search speaks the same query language as the website's Advanced Search:
field-scoped Lucene over a Solr index. `mediatype:texts`, `subject:mathematics`,
`date:[2010-01-01 TO 2012-12-31]`, and free text all work. Results come back as
a stream of documents you can sort, project, and page through.

## The Wayback Machine is separate

Web captures live in the **Wayback Machine**, a different service with its own
APIs: an availability lookup, the CDX capture-history server, and Save Page Now.
archive folds them into the `wayback` command group. A Wayback URL is addressed
by the original URL plus a timestamp, not by an item identifier.

```bash
archive wayback available example.com
archive wayback get example.com -t 2010 --text
```

## Anonymous by default

Reading public data needs no account. Credentials (an IAS3 access/secret pair
from your [archive.org account](https://archive.org/account/s3.php)) are only
required to **upload**, **delete**, or read your **task** queue. See
[configuration](/reference/configuration/) for how to store them.

## Output is yours to shape

Every command renders through one output layer, so the same data is a table for
reading, JSON or JSONL for piping, CSV/TSV for a spreadsheet, or a bare list of
URLs or identifiers for `xargs`. Pick with `-o`; project columns with
`--fields`. See [output formats](/reference/output/).
