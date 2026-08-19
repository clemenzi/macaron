# Service development

A service must contain a `.macaron` directory with at least a `start` script:

```text
service-name/
└── .macaron/
    ├── build       # optional
    ├── cleanup     # optional
    ├── doctor      # optional
    └── start       # required
```

Scripts can be executable files or Bash scripts run by Macaron. Every script
runs with the service repository root as its working directory.

## The `start` script

Macaron looks for `.macaron/start` in every service and starts them in
alphabetical order without waiting for one service to finish before launching
the next. Each service is assigned a free TCP port, starting at `49001`,
through the `MACARON_AVAILABLE_PORT` environment variable.

The script must keep the service in the foreground; normally it should `exec`
the server so Macaron can track it. Macaron owns that process and terminates it
when Macaron stops. For example:

```bash
#!/usr/bin/env bash
set -e

exec ./server --host 0.0.0.0 --port "$MACARON_AVAILABLE_PORT"
```

Macaron displays the service URL using the Mac's Tailscale IPv4 address and the
assigned port. Once startup completes, successfully started services and their
ports are also available in `~/.config/macaron/active-services.json`; the file
is removed when Macaron stops.

## The `build` script

If present, `.macaron/build` is offered during `macaron install`. It can be
used to install dependencies or build the service.

## The `doctor` script

If present, `.macaron/doctor` is run after the build and by `macaron doctor`. It
must exit with code `0` when the service is configured correctly and with a
non-zero code when an error occurs.

## The `cleanup` script

If present, `.macaron/cleanup` is run when `macaron` stops, after Macaron has
terminated the service process and before it restores the previous system
settings. Use it only for additional cleanup, such as removing temporary
resources created by the service. Cleanup scripts run for every enabled
service; a failing script is reported but does not prevent the other services
or Macaron itself from being cleaned up.
