# Macaron

Macaron turns your Mac into a remotely accessible workstation. It uses
[Tailscale][tailscale] to connect to your private network, keeps your Mac
awake, and starts the services you have installed.

## Requirements

- macOS;
- [Tailscale][tailscale] installed and authenticated;
- `sudo` access.

## Installation

Install Macaron with:

```sh
curl -L https://raw.githubusercontent.com/clemenzi/macaron/refs/heads/main/install.sh | sudo bash
```

This installs `macaron` in `/usr/local/bin`.

## Starting Macaron

Start Macaron with:

```sh
macaron
```

When it starts, Macaron enables Remote Login (SSH), prevents your Mac from
going to sleep, starts Tailscale, and loads the configured services. It then
prints the SSH address and the URLs available over the Tailscale network.

Press `Ctrl-C` to stop it. Remote Login, sleep, and Tailscale are restored to
their previous state.

## Service management

Macaron does not include any default services. You can install a service from a
Git repository with:

```sh
macaron install <repository>
```

The most useful commands are:

```sh
macaron list                  # List installed services
macaron update                # Update services
macaron disable <service>     # Disable a service
macaron enable <service>      # Enable a service
macaron delete <service>      # Remove a service
macaron doctor                # Check the configuration
macaron self-update           # Update macaron itself
```

For the complete documentation, see:

- [service management](docs/services.md), for installing and using services;
- [service development](docs/service-development.md), for creating a service;
- [advanced commands](docs/commands.md), for all available options.

> [!NOTE]
> Services run while Macaron is running. Macaron tracks and stops their
> foreground service processes when it exits.

## License

[MIT](LICENSE)

[tailscale]: https://tailscale.com/
