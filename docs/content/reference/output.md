---
title: "Output formats"
description: "The seven output formats, field projection, and per-row templates."
weight: 30
---

Every command renders through one output layer, so the same data is a table for
reading or a stream for piping. Choose with `-o`/`--output`.

## The formats

| `-o` value | What you get |
|-----------|--------------|
| `table` | aligned, human-readable columns (the default on a terminal) |
| `json` | one JSON array of objects |
| `jsonl` | one JSON object per line (newline-delimited) |
| `csv` | comma-separated, with a header row |
| `tsv` | tab-separated, with a header row |
| `url` | one URL per row (details, download, or replay, by command) |
| `raw` | the bare primary value per row (e.g. an identifier or file bytes) |

## auto

The default, `auto`, picks `table` when stdout is a terminal and `jsonl` when
it is a pipe or file, so interactive use is readable and scripted use is
machine-friendly without a flag:

```bash
archive search 'collection:nasa' -n 3            # table on screen
archive search 'collection:nasa' -n 3 | cat       # jsonl into a pipe
```

## Projecting columns

`--fields` restricts and orders the columns by name, across every format:

```bash
archive search 'collection:nasa' --fields identifier,downloads
archive search 'collection:nasa' --fields identifier,downloads -o csv
archive files nasa --fields name,size -o tsv
```

`--no-header` drops the header row from `table`, `csv`, and `tsv` output, which
is handy when feeding a column straight into another tool.

## URLs and raw values for pipelines

`-o url` gives the natural URL for each row, and `-o raw` gives the bare primary
value, both newline-delimited and perfect for `xargs`:

```bash
archive search 'collection:nasa' -o url | head
archive search 'collection:nasa' --fields identifier -o raw -n 5 | xargs -n1 archive item
archive files nasa --glob '*.jpg' -o url | xargs -n1 curl -sO
```

For `wayback get`, `-o raw` (the default) streams the snapshot bytes straight to
stdout, so you can redirect them to a file or pipe them onward.

## Templates

For full control, `--template` applies a Go `text/template` to each row's value
object and prints the result:

```bash
archive search 'collection:nasa' -n 5 \
  --template '{{.identifier}} has {{.downloads}} downloads'
```

The template sees the same fields you would get from `-o json`, so inspect a row
with `-o json` first to learn the field names, then template over them.
