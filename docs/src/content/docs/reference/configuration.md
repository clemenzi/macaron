---
title: Files and runtime state
description: Reference for Macaron directories, environment variables, ports, and generated state.
---

## Configuration directory

Macaron follows `XDG_CONFIG_HOME`. Its base directory is:

```text
$XDG_CONFIG_HOME/macaron/
```

When `XDG_CONFIG_HOME` is not set, it defaults to:

```text
~/.config/macaron/
```

| Path | Purpose |
| --- | --- |
| `services/` | Enabled services |
| `services-disabled/` | Disabled services |
| `active-services.json` | Services that survived the startup check and their ports |

## Active-service file

While Macaron is running, `active-services.json` contains an array like:

```json
[
  { "name": "dashboard", "port": 49001 },
  { "name": "notes", "port": 49002 }
]
```

The file is written atomically after startup and removed during shutdown. Services that exit within the one-second startup window are not included. Treat this as ephemeral runtime state, not persistent configuration.

## Port allocation

Macaron checks TCP ports in ascending order starting at `49001`. Occupied ports are skipped. Each start script receives its selected port through:

```text
MACARON_AVAILABLE_PORT
```

This variable is provided only to the service process. Macaron does not configure the application's listen address or protocol.

## System state

At startup Macaron records whether Remote Login is enabled, whether sleep is already disabled, and whether Tailscale is active. It then enables all three capabilities. During shutdown it attempts to restore each recorded value even if service cleanup reports an error.

Macaron does not modify SSH keys, Tailscale access-control rules, firewall rules, or application configuration.
