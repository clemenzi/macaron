---
title: Troubleshooting
description: Diagnose Tailscale, startup, service, and cleanup problems in Macaron.
---

## Start with doctor

```sh
macaron doctor
```

Doctor verifies that the `tailscale` command exists, reads its JSON status, detects a logged-out state, and runs `.macaron/doctor` for every enabled service that provides one. It exits with code `1` if any check fails.

## Tailscale is not found

Ensure the Tailscale CLI is installed and visible in the shell that launches Macaron:

```sh
command -v tailscale
tailscale status
```

If Tailscale reports `NeedsLogin`, authenticate it before retrying.

## No services are found

Check the enabled service directory and the required start file:

```sh
macaron list
ls -la ~/.config/macaron/services/<service>/.macaron/start
```

Only regular `.macaron/start` files are launched. Disabled services are intentionally ignored.

## A service exits immediately

Run the start script from the service root and provide a test port:

```sh
cd ~/.config/macaron/services/<service>
MACARON_AVAILABLE_PORT=49001 bash .macaron/start
```

The script must keep the application in the foreground. Remove backgrounding or daemon flags and use `exec` for the final server command.

## The URL cannot be reached

Confirm that:

- the client and Mac are connected to the same Tailscale network;
- the Tailscale access policy allows the connection;
- the application uses `MACARON_AVAILABLE_PORT`;
- the application listens on `0.0.0.0` or another Tailscale-reachable address, not only `127.0.0.1`;
- no application-level authentication or firewall rule is rejecting the request.

Use `active-services.json` to confirm the port Macaron assigned.

## Settings were not restored

Macaron restores state during its normal `SIGINT` and `SIGTERM` shutdown path. A forced kill, system crash, or power loss cannot run cleanup. In that case, restore the relevant setting manually and restart Tailscale as needed:

```sh
sudo systemsetup -setremotelogin off
sudo pmset -a disablesleep 0
tailscale down
```

Only change values that match the state you want for your Mac.
