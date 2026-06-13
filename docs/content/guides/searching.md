---
title: "Searching"
description: "Build Lucene queries, sort and project fields, filter by mediatype and date, and export large result sets."
weight: 10
---

`archive search` queries the Internet Archive's Advanced Search (Solr) index. A
query is Lucene: free text, field-scoped terms, ranges, and boolean operators.

```bash
archive search 'collection:nasa' -n 5
archive search 'mediatype:texts AND subject:mathematics' -n 5
archive search 'title:(apollo AND moon)' -n 5
```

## Field filters

Common fields have dedicated flags so you do not have to hand-write the Lucene:

```bash
archive search apollo --media image -n 5
archive search apollo --collection nasa -n 5
archive search apollo --creator 'NASA' -n 5
archive search apollo --year 1969 -n 5
archive search apollo --year 1965-1972 -n 5
```

These compose with a free-text or field query; everything is `AND`ed together.

## Sorting

`--sort` takes a Solr sort key and is repeatable for tie-breaks:

```bash
archive search 'collection:nasa' --sort 'downloads desc' -n 5
archive search 'collection:nasa' --sort 'date desc' --sort 'downloads desc' -n 5
```

## Choosing the columns

By default each result carries identifier, title, mediatype, downloads, and
date. Ask for specific metadata fields with `-f` (repeatable), and reshape the
displayed columns with `--fields`:

```bash
archive search 'collection:nasa' -f identifier -f downloads -f publicdate
archive search 'collection:nasa' --fields identifier,downloads -o csv
```

## Just the count

```bash
archive search 'collection:nasa AND mediatype:image' --count
```

## Output that pipes

```bash
archive search 'collection:nasa' -o url            # details URLs
archive search 'collection:nasa' -o jsonl          # one JSON object per line
archive search 'collection:nasa' --fields identifier -o url | head
```

Pipe identifiers straight into another command:

```bash
archive search 'collection:nasa AND mediatype:image' --fields identifier -o raw -n 10 \
  | xargs -n1 archive files
```

## Large result sets

A normal search pages through the Advanced Search endpoint up to the `--limit`
you set (`-n 0` means unlimited, bounded only by what Solr will return). To
export a very large set reliably, `--all` switches to the cursor-based Scraping
API, which has no deep-paging ceiling:

```bash
archive search 'collection:nasa' --all --fields identifier -o raw > nasa-ids.txt
```

Tune the per-request page size with `--rows`. The client rate-limits and
retries on 429/5xx automatically, so a long export stays polite.
