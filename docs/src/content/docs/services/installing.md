---
title: Install a service
description: Add services from Git repositories or local directories and control setup checks.
---

## Install from Git

Pass any source accepted by `git clone`:

```sh
macaron install https://github.com/example/macaron-service-dashboard.git
```

The destination name comes from the repository name. A leading `macaron-service-` prefix is removed, so the example above becomes `dashboard`.

Choose a branch or an explicit service name when needed:

```sh
macaron install --branch develop --name dashboard \
  https://github.com/example/dashboard.git
```

## Install from a local directory

```sh
macaron install ./my-service
```

Macaron copies the directory, including hidden files and symbolic links. The source does not need to be a Git repository. `--branch` cannot be used with a local directory.

## Build and validate

If `.macaron/build` exists, Macaron asks before running it. Use `--yes` for unattended installation or `--skip-build` to bypass it:

```sh
macaron install --yes <source>
macaron install --skip-build <source>
```

If `.macaron/doctor` exists, it runs after the build. A failed doctor check is reported but does not remove the installed service. Skip it with `--skip-doctor`.

:::caution
Build scripts come from the service source and execute as your user. Review a service before approving its build script.
:::

## Naming and conflicts

A service name must be a single, non-empty path component. Names containing `/` or `\\`, and the names `.` and `..`, are rejected. Installation also stops if an enabled service with the same destination name already exists.
