# Lattice

A modular terminal dashboard built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Lattice ships with a set of built-in modules and supports external plugins written in any language.

```
╭──────────────────────╮╭──────────────────────╮
│ LATTICE              ││ CLOCK                │
│ Good evening, user   ││      20:31:47        │
│ Your terminal dash   ││  Sunday, Mar 16      │
╰──────────────────────╯╰──────────────────────╯
╭──────────────────────╮╭──────────────────────╮
│ SYSTEM               ││ GITHUB               │
│ CPU  12% ██░░░░░░░░  ││ COMMITS          4   │
│ MEM  58% █████░░░░░  ││ PRS MERGED       1   │
╰──────────────────────╯╰──────────────────────╯
╭──────────────────────╮╭──────────────────────╮
│ WEATHER              ││ UPTIME               │
│ ☀ Clear              ││ 3d 7h 22m            │
│ +18°C (feels +16)    ││ since Mar 13, 08:14  │
╰──────────────────────╯╰──────────────────────╯
```

## Install

```bash
go install github.com/floatpane/lattice@latest
```

Or build from source:

```bash
git clone https://github.com/floatpane/lattice.git
cd lattice
go build -o lattice .
```

## Usage

```bash
lattice              # launch the dashboard
lattice list         # show available modules
lattice import <pkg> # install an external plugin
lattice remove <name> # remove an installed plugin
lattice help         # show help
```

Press `q` or `Ctrl+C` to quit the dashboard.

## Configuration

Lattice reads its config from `~/.config/lattice/config.yaml`. If the file doesn't exist, it uses a default set of modules.

Copy the example config to get started:

```bash
mkdir -p ~/.config/lattice
cp config.example.yaml ~/.config/lattice/config.yaml
```

### Config format

```yaml
# Number of columns in the grid layout
columns: 2

# Modules to display, in order
modules:
  - type: greeting
    config:
      name: "Ada"

  - type: clock
  - type: system

  - type: github
    config:
      username: "octocat"
      token: "ghp_..."

  - type: weather
    config:
      city: "London"

  - type: uptime
```

Modules are arranged in columns using round-robin distribution. The order in the config determines the order on screen.

### Environment variables

Secrets and config values can also be set via environment variables (or a `.env` file in the working directory). Env vars take effect when the config key is not set:

| Module   | Config key | Env var            |
|----------|------------|--------------------|
| greeting | `name`     | `LATTICE_NAME`     |
| github   | `username` | `GITHUB_USERNAME`  |
| github   | `token`    | `GITHUB_TOKEN`     |
| weather  | `city`     | `LATTICE_CITY`     |

## Built-in modules

| Module     | Description                                |
|------------|--------------------------------------------|
| `greeting` | Time-aware greeting with your name         |
| `clock`    | Live clock with date                       |
| `system`   | CPU, memory, and GPU usage bars            |
| `github`   | Today's commits, merged PRs, closed issues |
| `weather`  | Current weather via [wttr.in](https://wttr.in) (no API key needed) |
| `uptime`   | System uptime since last boot              |

## Plugins

Lattice supports external plugins — standalone binaries that communicate over JSON stdin/stdout. No recompilation needed.

### Installing a plugin

```bash
# Install a Go-based plugin
lattice import github.com/someone/lattice-spotify@latest

# Then add it to your config
# modules:
#   - type: spotify
```

Plugins are installed to `~/.config/lattice/plugins/`. Lattice also searches your `$PATH` for binaries named `lattice-<name>`.

### Removing a plugin

```bash
lattice remove spotify
```

## Creating plugins

See [DEVELOPING.md](DEVELOPING.md) for the full plugin development guide.

### Quick version

A plugin is any executable named `lattice-<name>` that reads newline-delimited JSON from stdin and writes JSON responses to stdout.

**Go (using the SDK):**

```go
package main

import "github.com/floatpane/lattice/pkg/plugin"

func main() {
    plugin.Run(func(req plugin.Request) plugin.Response {
        switch req.Type {
        case "init":
            return plugin.Response{
                Name:     "MY MODULE",
                Interval: 30,
            }
        case "update", "view":
            return plugin.Response{
                Content: "Hello from my plugin!",
            }
        }
        return plugin.Response{}
    })
}
```

**Python:**

```python
#!/usr/bin/env python3
import json, sys

for line in sys.stdin:
    req = json.loads(line)
    if req["type"] == "init":
        print(json.dumps({"name": "PYMOD", "interval": 10}), flush=True)
    elif req["type"] in ("update", "view"):
        print(json.dumps({"content": "Hello from Python!"}), flush=True)
```

**Bash:**

```bash
#!/usr/bin/env bash
while IFS= read -r line; do
  type=$(echo "$line" | jq -r .type)
  case "$type" in
    init)   echo '{"name":"SHMOD","interval":60}' ;;
    *)      echo '{"content":"Hello from bash!"}' ;;
  esac
done
```

## License

The software is protected by the MIT License. See [LICENSE](LICENSE) for details.
