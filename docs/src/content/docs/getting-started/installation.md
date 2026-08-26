---
title: Installation
description: Requirements, installer behavior, updates, and uninstall guidance for Macaron.
---

## Requirements

Macaron requires:

- a Mac with Apple Silicon or an Intel processor;
- [Tailscale](https://tailscale.com/download/mac) installed, available on `PATH`, and signed in;
- an account allowed to use `sudo`;
- Git when you want to install or update services from repositories.

Run a configuration check before your first session:

```sh
macaron doctor
```

## Install the binary

```sh
curl -L https://raw.githubusercontent.com/clemenzi/macaron/refs/heads/main/install.sh | sudo bash
```

The installer detects `arm64` or `x86_64`, downloads the matching binary from the latest GitHub release, and installs it as:

```text
/usr/local/bin/macaron
```

It also creates the enabled and disabled service directories for the user who invoked `sudo`. Go is not required on the destination Mac.

:::note
The command downloads and runs a shell script with administrator privileges. Review [`install.sh`](https://github.com/clemenzi/macaron/blob/main/install.sh) first if that does not match your security policy.
:::

## Update Macaron

Use the built-in updater:

```sh
macaron self-update
```

This downloads the current installer to a temporary file and runs it with `sudo`. Running the installation command again has the same effect.

## Uninstall

Macaron does not currently provide an uninstall command. Remove the binary manually if you no longer need it:

```sh
sudo rm /usr/local/bin/macaron
```

Service data is separate and remains in `~/.config/macaron/` unless you remove it yourself.
