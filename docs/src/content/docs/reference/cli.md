---
title: CLI reference
description: Complete Macaron command syntax, install options, and exit behavior.
---

## Commands

| Command | Description |
| --- | --- |
| `macaron` | Start Macaron |
| `macaron start` | Start Macaron explicitly |
| `macaron doctor` | Check Tailscale and enabled services |
| `macaron install [options] <source>` | Install a service |
| `macaron update` | Pull and rebuild enabled Git services |
| `macaron disable <service>` | Move a service to the disabled directory |
| `macaron enable <service>` | Move a service to the enabled directory |
| `macaron delete <service>` | Permanently remove a service directory |
| `macaron list` | List enabled and disabled services |
| `macaron self-update` | Install the latest Macaron release |
| `macaron help` | Show command help |

## Install options

| Option | Description |
| --- | --- |
| `--name <name>` | Set the destination service name |
| `--branch <branch>` | Clone a specific Git branch |
| `--skip-build` | Do not run `.macaron/build` |
| `--skip-doctor` | Do not run `.macaron/doctor` after installation |
| `-y`, `--yes` | Approve the build script without prompting |
| `-h`, `--help` | Show install help |

Options can use short-option handling where a short alias exists. `install` accepts exactly one source; service state commands accept exactly one service name.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Command completed successfully |
| `1` | Runtime, dependency, service, or health-check failure |
| `2` | Invalid arguments, invalid service name, or unknown command |
