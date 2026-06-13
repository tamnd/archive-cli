---
title: "Installation"
description: "Install the archive binary with Homebrew, Scoop, a Linux package, Docker, go install, or a prebuilt archive."
weight: 20
---

archive is a single static binary with no runtime dependencies. Pick whichever
install path fits your platform.

## Homebrew (macOS, Linux)

```bash
brew install tamnd/tap/archive
```

## Scoop (Windows)

```powershell
scoop bucket add tamnd https://github.com/tamnd/scoop-bucket
scoop install archive
```

## Linux packages

Download the `.deb`, `.rpm`, or `.apk` for your architecture from the
[releases page](https://github.com/tamnd/archive-cli/releases) and install it:

```bash
sudo dpkg -i archive_*_linux_amd64.deb      # Debian / Ubuntu
sudo rpm -i  archive_*_linux_amd64.rpm       # Fedora / RHEL
sudo apk add --allow-untrusted archive_*.apk # Alpine
```

## Docker

```bash
docker run --rm -v ~/data/archive:/data ghcr.io/tamnd/archive search nasa -n 5
```

The image stores all state under `/data`; mount a volume to keep the cache and
your downloads across runs.

## go install

With a recent Go toolchain:

```bash
go install github.com/tamnd/archive-cli/cmd/archive@latest
```

## Prebuilt archives

Every release ships `tar.gz` (and `.zip` for Windows) archives for Linux,
macOS, Windows, and FreeBSD on amd64 and arm64. Each is accompanied by a
`checksums.txt`, a cosign signature, and a CycloneDX SBOM. Download, verify,
extract, and put the binary on your `PATH`:

```bash
tar xzf archive_*_linux_amd64.tar.gz
sudo mv archive /usr/local/bin/
```

## Verify it

```bash
archive version
```

Then jump to the [quick start](/getting-started/quick-start/).
