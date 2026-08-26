---
title: Service structure
description: Understand the directory contract and runtime model for a Macaron service.
---

A service is an ordinary application directory containing a `.macaron` folder:

```text
service-name/
├── .macaron/
│   ├── build       # optional
│   ├── cleanup     # optional
│   ├── doctor      # optional
│   └── start       # required to run
└── ... application files
```

Only regular files are recognized as lifecycle scripts. A script can be executable with its own shebang, or non-executable; non-executable scripts are run with Bash. Every script uses the service root as its working directory.

## Minimal service

This example serves the current directory with Python:

```bash title=".macaron/start"
#!/usr/bin/env bash
set -euo pipefail

exec python3 -m http.server "$MACARON_AVAILABLE_PORT" --bind 0.0.0.0
```

Make the script executable if you want its shebang to choose the interpreter:

```sh
chmod +x .macaron/start
```

Install the directory and start Macaron:

```sh
macaron install ./service-name
macaron
```

## Runtime contract

Macaron provides `MACARON_AVAILABLE_PORT` to the start script. The value is a free TCP port, beginning at `49001`. Your application must use this value instead of a hard-coded port.

The start script must remain in the foreground. Prefer `exec` for the final command so signals reach the application directly and Macaron can track its real process:

```bash
exec ./server --host 0.0.0.0 --port "$MACARON_AVAILABLE_PORT"
```

Macaron waits one second after launching all services. A process that has already exited is treated as a startup failure and is omitted from the active-service list.

## Logs and shutdown

Standard output and standard error are streamed to the Macaron terminal with the service name. On shutdown, the process receives `SIGTERM`; it has up to one second to exit before Macaron kills it. Implement graceful `SIGTERM` handling when your service needs to flush data or close connections.
