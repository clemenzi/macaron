---
title: Official services
description: Browse and install the services maintained by the Macaron project.
---

Official services are maintained in the [`macaron-services`](https://github.com/macaron-services) GitHub organization. They follow the same `.macaron` contract as any custom service and can be inspected before installation.

| Service | Purpose | Repository |
| --- | --- | --- |
| code-server | Run VS Code in a browser | [`macaron-services/code-server`](https://github.com/macaron-services/code-server) |
| dufs | Browse and manage files from a browser | [`macaron-services/dufs`](https://github.com/macaron-services/dufs) |

## code-server

The code-server service provides a browser-based VS Code environment running on your Mac.

Install it with:

```sh
macaron install https://github.com/macaron-services/code-server
```

During installation, its build script checks whether `code-server` is available and offers to install it with Homebrew when missing. The service binds to the port selected by Macaron on all network interfaces.

:::caution
The official start script launches code-server with application authentication disabled. Access must be restricted through your trusted Tailscale network and its access-control policy.
:::

- [View the service source](https://github.com/macaron-services/code-server)
- [Learn more about code-server](https://github.com/coder/code-server)

## dufs

The dufs service exposes a directory through a web file manager. Upload, download, edit, search, archive, and delete operations are enabled.

Install it with:

```sh
macaron install https://github.com/macaron-services/dufs
```

By default, dufs serves the current user's home directory. Set `DUFS_SERVE_PATH` when starting Macaron to expose a narrower location:

```sh
DUFS_SERVE_PATH="$HOME/Downloads" macaron start
```

Its build script checks whether `dufs` is available and offers to install it with Homebrew when missing.

:::danger
This service enables all file operations and does not add application-level authentication. Use a trusted Tailscale network and choose a restricted `DUFS_SERVE_PATH` unless you intend to expose the entire home directory to authorized network members.
:::

- [View the service source](https://github.com/macaron-services/dufs)
- [Learn more about dufs](https://github.com/sigoden/dufs)

## Find new services

This page reflects the official organization at the time of the latest documentation update. Visit [`github.com/macaron-services`](https://github.com/macaron-services) for newly published services and inspect their `.macaron` scripts before installation.
