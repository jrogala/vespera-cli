# vespera-cli

CLI for Vaonis Vespera 2 telescope control and FTP image downloads.

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
# Check telescope status (auto-connect to Vespera WiFi)
vespera status --auto-wifi

# List all observation sessions
vespera list

# Download all files from the latest session
vespera download --latest

# Show directory tree of all observations
vespera tree

# Download a specific session with auto WiFi switching
vespera download --auto-wifi --session 2026-03-22_orion
```

## JSON Output

All commands support `--json` for machine-readable output.
