---
title: "Uploading"
description: "Push files into your own items over the S3-like IAS3 interface, set metadata, and delete files."
weight: 40
---

`archive upload` writes files into an item over the Internet Archive's S3-like
IAS3 interface. Unlike everything else in this tool, it needs credentials.

## Credentials first

Get an access/secret key pair from your account at
[archive.org/account/s3.php](https://archive.org/account/s3.php), then store it
once:

```bash
archive configure
```

This prompts for the keys and writes them to `~/.config/archive/credentials`
with `0600` permissions. You can also pass `--access`/`--secret` per command or
set `ARCHIVE_ACCESS_KEY` / `ARCHIVE_SECRET_KEY`. Confirm what is configured:

```bash
archive whoami
```

See [configuration](/reference/configuration/) for the full resolution order.

## Uploading files

```bash
archive upload my-item ./photo.jpg
archive upload my-item ./a.jpg ./b.jpg ./c.jpg
```

To create the item on first upload, pass `--make-bucket` and set the metadata
that describes it:

```bash
archive upload my-new-item ./paper.pdf \
  --make-bucket \
  -m title="My Paper" \
  -m mediatype=texts \
  -m 'subject=physics'
```

Each `-m key:value` (or `key=value`) becomes an `x-archive-meta-<key>` header on
the create request. Useful extras:

- `--name` sets the remote file name for a single-file upload.
- `--content-type` overrides the detected Content-Type.
- `--no-derive` skips the Archive's post-upload derivation step (useful when you
  will upload several files and want to derive once at the end).

## Previewing with --dry-run

Before sending anything, see exactly which request and headers archive would
make:

```bash
archive upload my-item ./paper.pdf --make-bucket -m title="My Paper" --dry-run
```

`--dry-run` prints the target URL and the `x-archive-*` headers and exits
without uploading, so you can check a batch before it runs.

## Deleting files

```bash
archive delete my-item old.jpg
archive delete my-item a.jpg b.jpg
```

`delete` asks for confirmation first; pass `-y` to skip the prompt in a script.
It also needs credentials and only works on items you control.
