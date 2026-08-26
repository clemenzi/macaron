---
title: Manage services
description: List, update, disable, enable, and delete installed Macaron services.
---

## List services

```sh
macaron list
```

Enabled and disabled services are shown separately. Enabled services live in `services/`; disabled services live in `services-disabled/`.

## Update services

```sh
macaron update
```

Macaron runs `git pull` for every enabled Git service. After a successful pull, it prints the short commit revision and runs `.macaron/build` automatically when present. Local directories that are not Git worktrees are skipped. Disabled services are not updated.

The update command stops at the first pull or build failure, so resolve that error before running it again.

## Disable and enable

```sh
macaron disable dashboard
macaron enable dashboard
```

Disabling moves the service directory rather than changing a metadata flag. Disabled services are excluded from startup, updates, doctor checks, and cleanup. Repeating either command when the service is already in the requested state is safe.

The operation fails if a service with the same name already exists at the destination.

## Delete

```sh
macaron delete dashboard
```

Delete accepts the service directory name, not its repository URL, and can remove either an enabled or disabled service. If the name exists in both locations, Macaron refuses to choose between them. Deleting a missing service is a no-op.

:::danger
Deletion recursively removes the service directory. Macaron does not keep a backup.
:::
