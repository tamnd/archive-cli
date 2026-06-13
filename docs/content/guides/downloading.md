---
title: "Downloading"
description: "Pull whole items or selected files with concurrency, md5 verification, resume, and a flat layout."
weight: 30
---

`archive download` fetches files from an item into a per-item directory. It runs
several files at once, skips files already present with a matching md5, and can
verify and resume.

## Whole item or selected files

```bash
archive download nasa                          # every file in the item
archive download nasa globe_west_540.jpg        # one named file
archive download nasa a.jpg b.jpg               # several named files
```

Files land under `<out-dir>/<identifier>/` so multiple items never collide.
Point `--out-dir` (`-d`) wherever you like:

```bash
archive download nasa -d ./downloads
```

The default destination is `~/data/archive/download` (override the whole data
root with `--data-dir` or the `ARCHIVE_DATA_DIR` environment variable).

## Filtering what to fetch

The same `--glob` and `--format` filters as `files` decide what to download:

```bash
archive download nasa --glob '*.jpg' -d .
archive download principleofrelat00eins --format PDF -d books
```

## Verification and resume

Files already on disk with a matching md5 are skipped, so re-running a download
only fetches what is missing or changed. Add `--verify` to re-check the md5 of
each file after it is fetched and fail loudly on a mismatch:

```bash
archive download nasa --format JPEG -d . --verify
```

An interrupted transfer leaves a `.part` file and resumes from where it stopped
on the next run, using an HTTP range request.

## Layout

Some items store files in sub-directories. `--flat` drops those path
components so everything lands directly in the item directory:

```bash
archive download some-item --flat -d .
```

## Concurrency and politeness

`-j` (default 8) sets how many files download at once; `--rate` sets the
minimum delay between requests, and `--retries` the number of backoff retries
on 429/5xx. The defaults are tuned to be fast without hammering the Archive.

## Streaming a single file to stdout

To pipe one file straight into another tool instead of saving it, use the
download URL with `-o raw` is not the path here; for a single file the simplest
route is `files ... -o url` piped to `curl`, or name the file and a destination
of `.`:

```bash
archive files nasa --glob '*.jpg' -o url | head -1 | xargs curl -s | wc -c
```
