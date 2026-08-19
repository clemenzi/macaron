# Service management

Services are directories installed by Macaron in:

```text
~/.config/macaron/services/
```

## Installing a service

```sh
macaron install <source>
```

`<source>` can be a Git URL, any path supported by `git clone`, or a local
directory. Local directories do not need to be Git repositories and are copied
into Macaron's services directory, including hidden files. The directory name
is derived from the source. To set it manually:

```sh
macaron install --name <name> <source>
```

During installation, if the source contains a `.macaron/build` script,
Macaron asks for confirmation before running it.

## Updating, disabling, and removing services

```sh
macaron update
macaron disable <service-name>
macaron enable <service-name>
macaron delete <service-name>
macaron list
```

After updating a Git service, Macaron automatically runs its `.macaron/build`
script when present and reports this to the user. No confirmation is requested.
Services installed from local directories that are not Git repositories are
skipped by `macaron update`.

`delete` takes the service directory name, not the repository URL.
It removes either an enabled or disabled service. If a service with the same
name is present in both locations, Macaron refuses to delete either copy.

`disable` moves a service from `~/.config/macaron/services/` to
`~/.config/macaron/services-disabled/`, so Macaron will not start, update, or
check it. Use `enable` to move it back. `list` shows enabled and disabled
services separately.

## Active services

While Macaron is running, it writes the services that started successfully and
their assigned ports to:

```text
~/.config/macaron/active-services.json
```

For example:

```json
[
  {"name":"example-service","port":49001}
]
```

The file is replaced atomically after startup and removed when Macaron stops.
Services that fail during startup are not included.

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
