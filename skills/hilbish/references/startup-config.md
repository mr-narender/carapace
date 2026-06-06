# Hilbish Startup & Configuration

In-depth reference for hilbish's initialization sequence, configuration files, XDG directory layout, shell options, prompt customization, and module loading.

## Overview

Hilbish follows an XDG-compliant directory layout and uses Lua for all configuration. The shell initializes in two phases: first the Go runtime sets up the Lua VM and readline, then the Lua `nature/` framework loads completions, commands, runners, and the user's `init.lua`.

## Source Files

| File | Language | Purpose |
|------|----------|---------|
| `main.go` | Go | Entry point, directory resolution, config loading |
| `lua.go` | Go | Lua VM initialization |
| `nature/init.lua` | Lua | Lua-side bootstrap, module loading, error handlers |
| `nature/completions/init.lua` | Lua | Auto-load per-command completions |
| `nature/runner.lua` | Lua | Runner mode setup |
| `nature/editor.lua` | Lua | Editor wrapper |
| `nature/vim.lua` | Lua | Vim mode setup |
| `nature/hilbish.lua` | Lua | `hilbish.run()`, deprecated `runnerMode` |
| `nature/opts/` | Lua | Shell options |
| `nature/commands/` | Lua | Built-in commands |
| `nature/processors/` | Lua | Command processors |

## Directory Layout

### XDG Directories

| Directory | Source | Default (Linux) | Purpose |
|-----------|--------|-----------------|---------|
| Config | `os.UserConfigDir()` | `~/.config/hilbish/` | User configuration |
| Data | `XDG_DATA_HOME` | `~/.local/share/hilbish/` | History, modules |
| System Data | `XDG_DATA_DIRS` | `/usr/share/hilbish/` | Default config, docs |

### Key Paths

| Path | Variable | Description |
|------|----------|-------------|
| `~/.config/hilbish/init.lua` | `defaultConfPath` | User configuration script |
| `~/.local/share/hilbish/.hilbish-history` | `defaultHistPath` | Command history file |
| `/usr/share/hilbish/.hilbishrc.lua` | `dataDir + "/.hilbishrc.lua"` | Sample/default configuration |
| `~/.local/share/hilbish/start/` | — | Auto-loaded startup modules |

### Go Directory Resolution

```go
// main.go
curuser, _ = user.Current()
confDir, _ = os.UserConfigDir()

switch runtime.GOOS {
case "linux", "darwin":
    userDataDir = getenv("XDG_DATA_HOME", curuser.HomeDir + "/.local/share")
default:
    userDataDir = confDir  // Windows: %APPDATA%
}

defaultConfDir = filepath.Join(confDir, "hilbish")
defaultConfPath = filepath.Join(defaultConfDir, "init.lua")
defaultHistDir = filepath.Join(userDataDir, "hilbish")
defaultHistPath = filepath.Join(defaultHistDir, ".hilbish-history")
```

On Linux, if `dataDir` is empty at compile time, hilbish searches `XDG_DATA_DIRS` for the sample config:

```go
if dataDir == "" {
    searchableDirs := getenv("XDG_DATA_DIRS", "/usr/local/share/:/usr/share/")
    for _, path := range strings.Split(searchableDirs, ":") {
        _, err := os.Stat(filepath.Join(path, "hilbish", ".hilbishrc.lua"))
        if err == nil {
            dataDir = filepath.Join(path, "hilbish")
            break
        }
    }
}
```

## Initialization Sequence

### Phase 1: Go Runtime (main.go)

1. **Parse command-line flags** — `-c`, `-C`, `-l`, `-i`, `-n`, `-S`, `-v`, `-h`
2. **Resolve directories** — config, data, history paths
3. **Create readline instance** — `newLineReader("", false)`
4. **Initialize Lua VM** — `luaInit()` creates `l *rt.Runtime`
5. **Start signal handler** — `go handleSignals()`
6. **Load configuration** — `runConfig(defaultConfPath)` or sample config
7. **Emit `hilbish.init`** — `hooks.Emit("hilbish.init")`
8. **Enter interactive loop** — or execute `-c` command / script file

### Phase 2: Lua Bootstrap (nature/init.lua)

```lua
-- Load function extensions
local _ = require 'succulent'

-- Extend package.path for hilbish modules
package.path = package.path .. ';'
    .. hilbish.dataDir .. '/?/init.lua' .. ';'
    .. hilbish.dataDir .. '/?/?.lua' .. ';'
    .. hilbish.dataDir .. '/?.lua'

-- Set up native module paths
hilbish.module.paths = '?.so;?/?.so;'
    .. hilbish.userDir.data .. 'hilbish/libs/?/?.so'
    .. ";" .. hilbish.userDir.data .. 'hilbish/libs/?.so'

-- Add custom module searcher
table.insert(package.searchers, function(module)
    local path = package.searchpath(module, hilbish.module.paths)
    if not path then return nil end
    return function() return hilbish.module.load(path) end, path
end)

-- Load nature modules
require 'nature.hilbish'
require 'nature.processors'
require 'nature.processors.wildcardWarn'
require 'nature.commands'
require 'nature.completions'
require 'nature.opts'
require 'nature.vim'
require 'nature.runner'
require 'nature.hummingbird'
require 'nature.env'
require 'nature.abbr'
require 'nature.editor'

-- Set SHLVL
local shlvl = tonumber(os.getenv 'SHLVL')
if shlvl ~= nil then
    os.setenv('SHLVL', tostring(shlvl + 1))
else
    os.setenv('SHLVL', '0')
end

-- Load startup modules
local startSearchPath = hilbish.userDir.data .. '/hilbish/start/?/init.lua;'
    .. hilbish.userDir.data .. '/hilbish/start/?.lua'
local ok, modules = pcall(fs.readdir, hilbish.userDir.data .. '/hilbish/start/')
if ok then
    for _, module in ipairs(modules) do
        local entry = package.searchpath(module, startSearchPath)
        if entry then dofile(entry) end
    end
end
package.path = package.path .. ';' .. startSearchPath

-- Set up error handlers
bait.catch('error', function(event, handler, err)
    print(string.format('Encountered an error in %s handler\n%s', event, err:sub(8)))
end)
bait.catch('command.not-found', function(cmd)
    print(string.format('hilbish: %s not found', cmd))
end)
bait.catch('command.not-executable', function(cmd)
    print(string.format('hilbish: %s: not executable', cmd))
end)
```

### Module Loading Order

| Order | Module | Purpose |
|-------|--------|---------|
| 1 | `succulent` | Lua function extensions (table.filter, etc.) |
| 2 | `nature.hilbish` | `hilbish.run()`, `hilbish.runnerMode()` |
| 3 | `nature.processors` | Command preprocessing |
| 4 | `nature.processors.wildcardWarn` | Wildcard warning |
| 5 | `nature.commands` | Built-in shell commands |
| 6 | `nature.completions` | Completion system + auto-load per-command completions |
| 7 | `nature.opts` | Shell options |
| 8 | `nature.vim` | Vim mode and registers |
| 9 | `nature.runner` | Runner mode system |
| 10 | `nature.hummingbird` | HTTP/networking |
| 11 | `nature.env` | Environment variable management |
| 12 | `nature.abbr` | Abbreviations |
| 13 | `nature.editor` | Line editor wrapper |
| 14 | Startup modules | User's `~/.local/share/hilbish/start/` |

## Configuration File

### Location

`~/.config/hilbish/init.lua` (Linux/macOS), `%APPDATA%/hilbish/init.lua` (Windows)

### Fallback

If the user config doesn't exist, hilbish loads the sample config:
1. `.hilbishrc.lua` in the current directory (for development)
2. `/usr/share/hilbish/.hilbishrc.lua` (system sample)

### Sample Configuration

```lua
local function doPrompt(fail)
    hilbish.prompt(lunacolors.format(
        '{blue}%u {cyan}%d ' .. (fail and '{red}' or '{green}') .. '∆ '
    ))
end

doPrompt()

bait.catch('command.exit', function(code)
    doPrompt(code ~= 0)
end)
```

### Common Configuration Patterns

```lua
-- Enable vim mode
hilbish.inputMode('vim')

-- Set aliases
hilbish.alias('ga', 'git add')
hilbish.alias('gs', 'git status')

-- Add to PATH
hilbish.appendPath {'~/go/bin', '~/.local/bin'}

-- Custom prompt with git branch
bait.catch('command.exit', function()
    local branch = hilbish.run('git rev-parse --abbrev-ref HEAD 2>/dev/null', false)
    branch = branch:gsub('\n', '')
    hilbish.prompt(lunacolors.format(
        '{green}%u {cyan}%d ' ..
        (branch ~= '' and '{yellow}(' .. branch .. ') ' or '') ..
        '{red}∆ '
    ))
end)

-- Custom greeting
hilbish.opts.greeting = 'Welcome to Hilbish!'

-- Enable fuzzy history search
hilbish.opts.fuzzy = true

-- Disable MOTD
hilbish.opts.motd = false

-- Custom syntax highlighting
function hilbish.highlighter(line)
    return line:gsub('"%w+"', function(c)
        return lunacolors.green(c)
    end)
end

-- Custom hinter
function hilbish.hinter(line, pos)
    if line:match('^cd ') then
        return 'directory'
    end
    return ''
end

-- Register custom command
local commander = require 'commander'
commander.register('hello', function(args, sinks)
    sinks.out:writeln('Hello ' .. (args[1] or 'world'))
end)

-- Register completion for custom command
hilbish.completions.add('command.hello', function(query, ctx, fields)
    local compGroup = {
        items = {'world', 'hilbish', 'lua'},
        type = 'grid'
    }
    return {compGroup}, query
end)
```

## Shell Options

The `hilbish.opts` table provides toggle/value options:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `autocd` | boolean | `false` | Enter directory when typed as command |
| `history` | boolean | `true` | Save command history |
| `greeting` | boolean\|string | `true` | Startup message (`false` to disable, or custom string) |
| `motd` | boolean | `true` | Show version/release info on startup |
| `fuzzy` | boolean | `false` | Fuzzy history search (Ctrl-R) |
| `notifyJobFinish` | boolean | `true` | Notify when background job finishes |
| `processorSkipList` | table | `{}` | Command processor names to skip |

## Prompt Customization

### Prompt Verbs

| Verb | Expansion |
|------|-----------|
| `%d` | Current working directory (abbreviated with `~`) |
| `%D` | Basename of current directory |
| `%h` | Hostname |
| `%u` | Username |

### Left and Right Prompts

```lua
hilbish.prompt '%u@%h :%d $'           -- left prompt
hilbish.prompt '%d', 'right'            -- right prompt
```

### Dynamic Prompt

Update the prompt on every command exit:

```lua
bait.catch('command.exit', function(code)
    local color = code ~= 0 and '{red}' or '{green}'
    hilbish.prompt(lunacolors.format(
        '{blue}%u {cyan}%d ' .. color .. '∆ '
    ))
end)
```

### Vim Mode Indicator

Show the current vim mode in the prompt:

```lua
bait.catch('hilbish.vimMode', function(mode)
    local modeColors = {
        normal = '{green}',
        insert = '{blue}',
        delete = '{red}',
        replace = '{yellow}'
    }
    local color = modeColors[mode] or '{white}'
    hilbish.prompt(lunacolors.format(
        color .. '[' .. mode .. '] {cyan}%d {green}∆ '
    ))
end)
```

### Multiline Prompt

If the prompt string contains `\n`, hilbish splits it:

```lua
hilbish.prompt 'user@host\n%d ∆ '
-- First line: "user@host" (static)
-- Second line: "~/directory ∆ " (editable)
```

### Continuation Prompt

```lua
hilbish.multiprompt '-->'
```

## Command-Line Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Print usage |
| `--version` | `-v` | Print version |
| `--command` | `-c` | Execute command |
| `--config` | `-C` | Set config file path |
| `--setshellenv` | `-S` | Set `$SHELL` to hilbish path |
| `--login` | `-l` | Force login shell |
| `--interactive` | `-i` | Force interactive shell |
| `--noexec` | `-n` | Don't execute (syntax check only) |

### Login Shell Detection

```go
if loginshflag || os.Args[0][0] == '-' {
    login = true
}
```

If the shell's argv[0] starts with `-` (e.g., `-hilbish`), it's treated as a login shell.

### Interactive Detection

```go
if *cmdflag == "" || interactiveflag {
    interactive = true
}
if fileInfo.Mode() & os.ModeCharDevice == 0 || !term.IsTerminal(int(os.Stdin.Fd())) {
    interactive = false
}
if getopt.NArgs() > 0 {
    interactive = false
}
```

Interactive mode is enabled when:
- No `-c` command and `-i` flag is set, OR
- stdin is a terminal and no script arguments

## Startup Modules

Hilbish auto-loads Lua modules from `~/.local/share/hilbish/start/`:

```lua
local startSearchPath = hilbish.userDir.data .. '/hilbish/start/?/init.lua;'
    .. hilbish.userDir.data .. '/hilbish/start/?.lua'

local ok, modules = pcall(fs.readdir, hilbish.userDir.data .. '/hilbish/start/')
if ok then
    for _, module in ipairs(modules) do
        local entry = package.searchpath(module, startSearchPath)
        if entry then dofile(entry) end
    end
end
```

This allows extending hilbish without modifying `init.lua` — just drop a module directory into `~/.local/share/hilbish/start/`.

## Setting Hilbish as Default Shell

### Not Recommended: Login Shell

```sh
chsh -s /usr/bin/hilbish
```

Hilbish is **not POSIX-compliant**, so some environment variables may be missing. This can break tools that expect a POSIX login shell.

### Recommended: Terminal Default

Configure hilbish as the default shell in your terminal emulator settings.

### Recommended: Run After Login Shell

Add to `~/.zlogin` or similar:

```sh
exec hilbish -S -l
```

The `-S` flag sets `$SHELL` to hilbish's path, and `-l` makes it a login shell.

## Edge Cases and Known Issues

- **No POSIX login shell compliance**: Hilbish does not source `/etc/profile`, `~/.profile`, or other POSIX startup files. Environment variables expected by some tools may be missing.
- **Config fallback is development-oriented**: The first fallback (`.hilbishrc.lua` in current directory) is meant for development, not production use.
- **`hilbish.opts.greeting` vs `motd`**: `greeting` controls the entire startup message, while `motd` specifically controls the version/release line. Setting `greeting = false` disables all startup output.
- **SHLVL starts at 0**: If `SHLVL` is not set in the environment, hilbish sets it to `0` (not `1` like some shells).
- **Startup modules use `dofile`**: Unlike `require`, `dofile` executes the file each time without caching. This means startup modules are always re-executed.
- **`succulent` is loaded globally**: The `require 'succulent'` call modifies global Lua types (adds `table.filter`, etc.). This cannot be disabled.

## References

- Hilbish getting started: <https://hilbish.sammyette.party/docs/getting-started>
- Hilbish install: <https://hilbish.sammyette.party/docs/install>
- Hilbish opts: <https://hilbish.sammyette.party/docs/features/opts>
- Source: `main.go` — Entry point, directory resolution, config loading
- Source: `nature/init.lua` — Lua bootstrap, module loading, error handlers

## Related Skills

- [Language & API](language.md) — The Lua API, runner mode, and bait event system
- [Completion](completion.md) — How completions are loaded and registered
- [Line Editing](line-editing.md) — Prompt customization and vim mode
- [Execution](execution.md) — How commands are executed after initialization
