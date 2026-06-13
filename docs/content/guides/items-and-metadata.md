---
title: "Items and metadata"
description: "Read an item's metadata record, a friendly summary, single fields, and its file listing."
weight: 20
---

Every item on archive.org has a metadata record and a set of files. archive
gives you three views of it: `item` for a summary, `metadata` for the raw
record, and `files` for the file listing.

## A friendly summary

```bash
archive item nasa
```

`item` pulls the identifier, title, mediatype, file count, total size, the
hosting server, and the details URL into one short table. It is the fastest way
to see what an identifier actually is.

## The raw metadata record

```bash
archive metadata nasa
```

This prints the full JSON document the Metadata API returns: the `metadata`
block, the `files` array, server and directory info, and timestamps. Pipe it
into `jq` for anything specific, or let archive pull a single field for you by
passing a slash-path subpath:

```bash
archive metadata nasa metadata/title       # just the title
archive metadata nasa metadata             # just the metadata block
archive metadata nasa files                 # just the files array
```

## Listing files

```bash
archive files nasa
```

Each row is one file: name, size, format, and md5. Narrow the listing with a
shell glob on the name or a substring match on archive's `format` field:

```bash
archive files nasa --glob '*.jpg'
archive files principleofrelat00eins --format PDF
```

The output composes like everything else:

```bash
archive files nasa --glob '*.jpg' -o url          # direct download URLs
archive files nasa --fields name,size -o csv       # a size report
```

## View counts and task history

Two more item-scoped commands round this out:

```bash
archive views nasa                  # all-time / 30-day / 7-day view counts
archive views nasa goody jeffm!     # several items at once
```

```bash
archive tasks nasa                  # the catalog / derive task history
```

`views` is anonymous. `tasks` needs [credentials](/reference/configuration/)
unless you own the item, because the task queue is private; without them the
Archive returns 401 and archive exits with code 4.
