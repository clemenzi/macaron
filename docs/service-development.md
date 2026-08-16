# Service development

A service must contain a `.macaron` directory with at least a `start` script:

```text
service-name/
└── .macaron/
    ├── build       # optional
    ├── doctor      # optional
    └── start       # required
```

Scripts can be executable files or Bash scripts run by Macaron.

## The `start` script

Macaron looks for `.macaron/start` in every service and runs them in
alphabetical order. Each service is assigned a port, starting at `49001`,
through the `MACARON_AVAILABLE_PORT` environment variable.

The script must start the service on the assigned port. For example:

```bash
#!/usr/bin/env bash
set -e

exec ./server --host 0.0.0.0 --port "$MACARON_AVAILABLE_PORT"
```

Macaron displays the service URL using the Mac's Tailscale IPv4 address and the
assigned port.

## The `build` script

If present, `.macaron/build` is offered during `macaron install`. It can be
used to install dependencies or build the service.

## The `doctor` script

If present, `.macaron/doctor` is run after the build and by `macaron doctor`. It
must exit with code `0` when the service is configured correctly and with a
non-zero code when an error occurs.
