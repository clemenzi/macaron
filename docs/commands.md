# Advanced commands

## General syntax

```text
macaron [start]
macaron doctor
macaron install [options] <source>
macaron update
macaron disable <service>
macaron enable <service>
macaron delete <service>
macaron list
macaron help
macaron self-update
```

## `install` options

| Option | Description |
| --- | --- |
| `--name <name>` | Set the service directory name |
| `--branch <branch>` | Clone a specific branch (Git repositories only) |
| `--skip-build` | Do not run the `.macaron/build` script |
| `--skip-doctor` | Do not run the `.macaron/doctor` script after installation |
| `-y`, `--yes` | Automatically confirm the build script |
| `-h`, `--help` | Show help for the `install` command |

Example:

```sh
macaron install --branch develop --yes https://github.com/example/service.git
```

`<source>` can also be a local directory. Its contents are copied into the
Macaron services directory:

```sh
macaron install ./my-service
```
