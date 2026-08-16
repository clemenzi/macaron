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

## Updating and removing services

```sh
macaron update
macaron delete <service-name>
macaron list
```

`delete` takes the service directory name, not the repository URL.

## Checking the configuration

```sh
macaron doctor
```

This command checks that Tailscale is installed and active and, when available,
runs each service's `.macaron/doctor` check. It returns a non-zero exit code if
any check fails.

For the structure required to create a service, see the [service development
guide](service-development.md).
