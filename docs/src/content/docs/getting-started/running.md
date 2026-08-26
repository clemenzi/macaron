---
title: Run and connect
description: Start Macaron, connect through Tailscale, and understand its shutdown behavior.
---

## Start a session

Both commands start Macaron:

```sh
macaron
# or
macaron start
```

Macaron first validates your `sudo` credentials, records the current system state, and then:

1. enables macOS Remote Login;
2. disables system sleep;
3. brings Tailscale up;
4. starts every enabled service that has a `.macaron/start` file;
5. prints the SSH command and each available service URL.

Services start alphabetically. Macaron assigns the first free TCP port at or above `49001` to each one.

## Connect remotely

The terminal prints an SSH command similar to:

```sh
ssh your-user@100.x.y.z
```

It also prints an HTTP URL for every service that remains running after its startup check:

```text
dashboard  http://100.x.y.z:49001
```

The remote device must belong to the same Tailscale network and be allowed by its access-control policy. A service must listen on the assigned port and on an address reachable through Tailscale, such as `0.0.0.0`.

## Stop safely

Press `Ctrl-C` or send `SIGTERM` to the Macaron process. During shutdown Macaron:

1. sends `SIGTERM` to each managed service;
2. force-kills services that do not stop within one second;
3. runs the cleanup script of every enabled service;
4. removes `active-services.json`;
5. restores the previous Remote Login, sleep, and Tailscale states.

Keep the terminal session running for as long as you need remote access. If Macaron exits because startup fails, its deferred cleanup still attempts to restore the captured system state.
