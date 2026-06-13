---
title: "Quick start"
description: "From an empty terminal to reading a real item's metadata and files, in a handful of commands."
weight: 30
---

This walks the core loop: search for items, look one up, and list its files.
Every command here hits live data on archive.org and finishes in a second or
two. None of it needs an account.

## 1. Search for items

```bash
archive search 'collection:nasa' -n 3
```

Each row is one item: its identifier, title, mediatype, and download count.
Pick the output that suits you:

```bash
archive search 'collection:nasa' -o url      # details URLs
archive search 'collection:nasa' -o table     # aligned columns for reading
archive search 'mediatype:texts AND subject:mathematics' -n 5
```

`search` speaks Lucene, the same query language as the website's Advanced
Search. Scope by field (`mediatype:`, `collection:`, `subject:`, `date:[...]`),
combine with `AND`/`OR`, and sort with `--sort`:

```bash
archive search 'collection:nasa' --sort 'downloads desc' -n 5
```

## 2. Look an item up

`item` prints a friendly summary; `metadata` prints the raw document.

```bash
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

Want a single field instead of the whole record?

```bash
archive metadata nasa metadata/title     # -> "NASA"
```

## 3. List the files

```bash
archive files nasa -o url
```

```
https://archive.org/download/nasa/globe_west_540.jpg
https://archive.org/download/nasa/globe_west_540_thumb.jpg
...
```

Filter by shell glob or by archive's `format` field:

```bash
archive files nasa --glob '*.jpg'
archive files nasa --format JPEG -o url
```

## 4. Download and verify

`download` fetches files concurrently into a per-item directory and verifies
each against its md5. Pass file names to select a subset, or a `--glob` /
`--format` filter.

```bash
archive download nasa globe_west_540_thumb.jpg -d .
```

```
ok   globe_west_540_thumb.jpg (6.0 KB)
```

A re-run skips files whose md5 already matches, and an interrupted download
resumes from where it stopped.

## 5. Travel the Wayback Machine

```bash
archive wayback available example.com         # the closest snapshot
archive wayback get example.com -t 2010 --text # the page's text, c. 2010
```

## Where to go next

- The [guides](/guides/) go deep on each task: searching, items and metadata,
  downloading, uploading, and the Wayback Machine.
- The [CLI reference](/reference/cli/) lists every command and flag.
- To upload or delete, set up [credentials](/reference/configuration/) first.
