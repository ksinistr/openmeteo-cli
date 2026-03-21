# Install `openmeteo-cli`

`openmeteo-cli` is a weather CLI built especially for AI agents. Its default `toon` output is compact and structured to minimize token usage while staying easy to read.

## Install For Humans And Agents

If you are using Clawbot or another coding agent, point it at this file:

```text
Install and configure openmeteo-cli by following the instructions here:
https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/install.md
```

## Quick Install

Install the latest release into `~/.local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/install.sh | bash
```

Install to a custom directory:

```bash
curl -fsSL https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/install.sh | INSTALL_DIR=/usr/local/bin bash
```

Use a directory you can write to. System paths such as `/usr/local/bin` may require elevated privileges.

Install a specific release:

```bash
curl -fsSL https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/install.sh | VERSION=v0.1.0 bash
```

## Verify

```bash
openmeteo-cli --help
openmeteo-cli today --lat 51.5074 --lon -0.1278
```

If the command is not found, add `~/.local/bin` to your shell profile:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Clawbot Skill

This repository includes a small skill for Clawbot in [skills/openmeteo-cli/SKILL.md](./skills/openmeteo-cli/SKILL.md).

Install the skill into your Clawbot skills directory:

```bash
curl -fsSL https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/skills/openmeteo-cli/install.sh | bash -s -- /path/to/clawbot/skills
```

That command creates:

```text
/path/to/clawbot/skills/openmeteo-cli/SKILL.md
```

After that, tell Clawbot to use the `openmeteo-cli` skill when it needs a weather forecast with compact `toon` output or machine-readable `json`.

If you already have the repository checked out locally, you can also run:

```bash
bash skills/openmeteo-cli/install.sh /path/to/clawbot/skills
```

## From Source

```bash
git clone https://github.com/ksinistr/openmeteo-cli.git
cd openmeteo-cli
make build
```

The local binary will be available at `bin/openmeteo-cli`.
