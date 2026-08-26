---
title: Lifecycle scripts
description: Define build, start, doctor, and cleanup behavior for a Macaron service.
---

All lifecycle scripts live in `.macaron/` and run from the service root.

## `start`

`start` is the only script required for a service to run. It receives `MACARON_AVAILABLE_PORT`, must start the long-lived application in the foreground, and should bind to an interface reachable through Tailscale.

```bash title=".macaron/start"
#!/usr/bin/env bash
set -euo pipefail

exec ./server --host 0.0.0.0 --port "$MACARON_AVAILABLE_PORT"
```

Do not append `&`, daemonize the process, or return after spawning a detached child. Macaron would consider the service stopped and could not shut it down reliably.

## `build`

Use `build` to install dependencies or compile the application:

```bash title=".macaron/build"
#!/usr/bin/env bash
set -euo pipefail

npm ci
npm run build
```

During installation, Macaron asks the user before running this script. During `macaron update`, it runs automatically after every successful Git pull. A build failure aborts installation or the remaining update sequence.

## `doctor`

Use `doctor` for configuration checks that can run without starting the service. Exit with `0` when healthy and a non-zero status when action is required:

```bash title=".macaron/doctor"
#!/usr/bin/env bash
set -euo pipefail

test -f .env || {
  echo "Missing .env"
  exit 1
}
```

It runs after installation unless skipped and whenever the user runs `macaron doctor`. Global doctor checks only enabled services.

## `cleanup`

`cleanup` runs after managed service processes have stopped and before Macaron restores system settings:

```bash title=".macaron/cleanup"
#!/usr/bin/env bash
set -euo pipefail

rm -f .runtime/socket
```

Use it for service-owned temporary resources, not for stopping the main application. Cleanup failures are reported but do not prevent other services or system restoration from being attempted.

## Script portability

- Declare required tools and environment in `doctor`.
- Use paths relative to the service root.
- Quote `"$MACARON_AVAILABLE_PORT"` and other variables.
- Use `set -euo pipefail` in Bash scripts when early failure is safer than partial setup.
- Keep scripts non-interactive after the installation build confirmation.
