# Service management

Services are Git repositories installed by Macaron in:

```text
~/.config/macaron/services/
```

## Installing a service

```sh
macaron install <repository>
```

`<repository>` can be a Git URL or any path supported by `git clone`. The
directory name is derived from the repository. To set it manually:

```sh
macaron install --name <name> <repository>
```

During installation, if the repository contains a `.macaron/build` script,
Macaron asks for confirmation before running it.

## Updating, disabling, and removing services

```sh
macaron update
macaron disable <service-name>
macaron enable <service-name>
macaron delete <service-name>
macaron list
```

After updating a service, Macaron automatically runs its `.macaron/build`
script when present and reports this to the user. No confirmation is requested.

`delete` takes the service directory name, not the repository URL.
It removes either an enabled or disabled service. If a service with the same
name is present in both locations, Macaron refuses to delete either copy.

`disable` moves a service from `~/.config/macaron/services/` to
`~/.config/macaron/services-disabled/`, so Macaron will not start, update, or
check it. Use `enable` to move it back. `list` shows enabled and disabled
services separately.

## Checking the configuration

```sh
macaron doctor
```

This command checks that Tailscale is installed and active and, when available,
runs each service's `.macaron/doctor` check. It returns a non-zero exit code if
any check fails.

Every `.macaron` script runs with its service repository as the working
directory.

For the structure required to create a service, see the [service development
guide](service-development.md).
