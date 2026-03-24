# vespera-cli

CLI for Vaonis Vespera 2 telescope control and FTP image downloads.

## Install

Download a binary from the [latest release](https://github.com/jrogala/vespera-cli/releases/latest), or install with Go:

```bash
go install github.com/jrogala/vespera-cli@latest
```

## Setup

Set `VESPERA_HOST` env var or use default (`10.0.0.1`):

```bash
export VESPERA_HOST=10.0.0.1
```

The `--auto-wifi` flag on any command will automatically switch to the Vespera WiFi network before executing.

## Commands

| Command | Description |
|---|---|
| `status` | Show telescope status and current observation |
| `list` | List observation sessions |
| `files` | List files in a session |
| `download` | Download images from the telescope via FTP |
| `tree` | Show full directory tree of stored observations |

## Examples

```bash
$ vespera status --auto-wifi
Connected to Vespera at 10.0.0.1
Observations: 42

$ vespera list
NAME                  DATE
2026-03-22_orion      2026-03-22
2026-03-21_polaris    2026-03-21
2026-03-20_m42        2026-03-20

$ vespera files 2026-03-22_orion
NAME                  TYPE  SIZE
2026-03-22_001.fits   FITS  15.3 MB
2026-03-22_002.fits   FITS  15.2 MB
2026-03-22_001.tiff   TIFF  45.8 MB
preview.jpg           JPEG  2.1 MB

$ vespera download --auto-wifi --session 2026-03-22_orion
Downloading 2026-03-22_orion...
Downloaded 4 files
```

## JSON Output

All commands support `--json` for machine-readable output.
