# Hilbish Language & API

In-depth reference for hilbish's Lua integration — the GopherLua runtime, the `hilbish` module API, the `bait` event system, runner mode, `commander`, `lunacolors`, and the `yarn` multi-threading library.

## Overview

Hilbish is "the Moon-powered shell" — it embeds a Lua runtime (GopherLua / `github.com/arnodel/golua`) at its core. All configuration, completion handlers, and extensibility flow through Lua. The shell provides a rich API surface via the `hilbish` module and its submodules.

## Source Files

| File | Language | Purpose |
|------|----------|---------|
| `api.go` | Go | Core `hilbish` module: exports map, `hilbishLoad()`, all `hl*` Go functions |
| `lua.go` | Go | Lua VM initialization (`l *rt.Runtime`) |
| `nature/hilbish.lua` | Lua | `hilbish.run()`, `hilbish.runnerMode()` (deprecated) |
| `nature/runner.lua` | Lua | Runner mode system: `hilbish.runner.add/set/get/exec/setCurrent` |
| `nature/editor.lua` | Lua | `hilbish.editor` wrapper |
| `nature/vim.lua` | Lua | `hilbish.vim` registers and mode |
| `nature/init.lua` | Lua | Bootstrap: requires all nature modules, sets up error handlers |
| `golibs/bait/` | Go | Bait event emitter module |
| `golibs/commander/` | Go | Commander module for Lua commands |
| `golibs/fs/` | Go | Filesystem module |
| `golibs/yarn/` | Go | Multi-threading module |
| `golibs/terminal/` | Go | Terminal control module |
| `libs/lunacolors/` | Lua | ANSI color/styling library |

## The GopherLua Runtime

Hilbish uses `github.com/arnodel/golua` (not `yuin/gopher-lua`) as its Lua runtime. Key characteristics:

- **Lua 5.3 compatible** — supports goto, bitwise operators, integer division `//`, and utf8 library
- **Go-embedded** — Go functions are exposed as Lua closures via `rt.NewGoFunction()`
- **Thread-safe** — the runtime supports multiple threads via `rt.Thread`
- **Global state** — `l *rt.Runtime` is the global Lua state, accessible from Go callbacks

### Module Loading

Hilbish extends Lua's `package.path` to include its data directory:

```lua
package.path = package.path .. ';'
    .. hilbish.dataDir .. '/?/init.lua' .. ';'
    .. hilbish.dataDir .. '/?/?.lua' .. ';'
    .. hilbish.dataDir .. '/?.lua'
```

Native (`.so`) modules are loaded via a custom searcher:

```lua
hilbish.module.paths = '?.so;?/?.so;'
    .. hilbish.userDir.data .. 'hilbish/libs/?/?.so'
    .. ';' .. hilbish.userDir.data .. 'hilbish/libs/?.so'

table.insert(package.searchers, function(module)
    local path = package.searchpath(module, hilbish.module.paths)
    if not path then return nil end
    return function() return hilbish.module.load(path) end, path
end)
```

## The `hilbish` Module

### Module Fields

| Field | Type | Description |
|-------|------|-------------|
| `ver` | string | Hilbish version |
| `goVersion` | string | Go version used to compile |
| `user` | string | Current username |
| `host` | string | Hostname |
| `home` | string | Home directory path |
| `dataDir` | string | Data directory (docs, default modules) |
| `interactive` | boolean | Whether running interactively |
| `login` | boolean | Whether running as login shell |
| `vimMode` | string\|nil | Current vim mode (nil in emacs mode) |
| `exitCode` | number | Exit code of last command |

### Submodules

| Submodule | Purpose |
|-----------|---------|
| `hilbish.completions` | Tab completion system |
| `hilbish.runner` | Runner mode management |
| `hilbish.aliases` | Command alias management |
| `hilbish.history` | Command history |
| `hilbish.jobs` | Job control |
| `hilbish.timers` | Timer management |
| `hilbish.editor` | Line editor access |
| `hilbish.vim` | Vim mode and registers |
| `hilbish.opts` | Shell options |
| `hilbish.userDir` | User directory paths |
| `hilbish.os` | OS-specific info |
| `hilbish.version` | Version details (branch, commit, release) |
| `hilbish.module` | Native module loading |
| `hilbish.sink` | I/O sink objects |

### Core Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `alias` | `alias(cmd, orig)` | Set command alias (supports `%1` substitutions) |
| `appendPath` | `appendPath(dir)` | Add directory to `$PATH` (string or table) |
| `cwd` | `cwd() → string` | Get current directory |
| `exec` | `exec(cmd)` | Replace shell with command (`syscall.Exec`) |
| `goro` | `goro(fn)` | Run function in goroutine (⚠️ may crash if accessing outside variables) |
| `highlighter` | `highlighter(line)` | Override for syntax highlighting |
| `hinter` | `hinter(line, pos)` | Override for inline hints |
| `inputMode` | `inputMode(mode)` | Set `"emacs"` or `"vim"` input mode |
| `interval` | `interval(cb, ms) → Timer` | Run callback every N milliseconds |
| `multiprompt` | `multiprompt(str)` | Set continuation prompt |
| `prependPath` | `prependPath(dir)` | Prepend directory to `$PATH` |
| `prompt` | `prompt(str, typ?)` | Set prompt (`"left"` or `"right"`) |
| `read` | `read(prompt?) → string` | Read user input (separate readline instance) |
| `run` | `run(cmd, streams?) → exitCode, stdout, stderr` | Execute shell command |
| `timeout` | `timeout(cb, ms) → Timer` | Run callback after N milliseconds |
| `which` | `which(name) → string` | Check if command exists, return path |

### Alias System

```lua
hilbish.alias('ga', 'git add')
hilbish.alias('dircount', 'ls %1 | wc -l')  -- numbered substitution
```

Numbered substitutions (`%1`, `%2`, etc.) replace positional arguments in the alias expansion.

### `hilbish.run(cmd, streams)`

Executes a shell command with configurable I/O:

```lua
-- Capture output
local exitCode, stdout, stderr = hilbish.run('ls -l', false)

-- Pipe commands
local fs = require 'fs'
local pr, pw = fs.pipe()
hilbish.run('ls -l', { stdout = pw, stderr = pw })
pw:close()
hilbish.run('wc -l', { stdin = pr })
```

The `streams` parameter:
- `boolean false` — capture output (returns exitCode, stdout, stderr)
- `table` — specify `{ stdout = sink, stderr = sink, stdin = sink }`
- `nil` or `true` — inherit stdio (returns exitCode only)

## The Bait Event System

Bait is hilbish's global event emitter, similar to Node.js's `EventEmitter`.

### API

| Function | Signature | Description |
|----------|-----------|-------------|
| `catch` | `catch(name, cb)` | Register event listener |
| `catchOnce` | `catchOnce(name, cb)` | Register one-time listener |
| `hooks` | `hooks(name) → table` | Get all listeners for an event |
| `release` | `release(name, catcher)` | Remove specific listener |
| `throw` | `throw(name, ...args)` | Emit event with arguments |

### Built-in Events

#### Command Events

| Event | Arguments | Description |
|-------|-----------|-------------|
| `command.preexec` | `input, cmdStr` | Before command execution |
| `command.exit` | `code, cmdStr` | After command completes |
| `command.not-found` | `cmdStr` | Command not found |
| `command.not-executable` | `cmdStr` | File not executable |
| `command.preprocess` | `input` | Before processing (alias expansion, etc.) |

#### Hilbish Events

| Event | Arguments | Description |
|-------|-----------|-------------|
| `hilbish.exit` | — | Shell is exiting |
| `hilbish.vimMode` | `modeName` | Vim mode changed |
| `hilbish.vimAction` | `actionName, args` | Vim action (yank/paste) |
| `hilbish.cancel` | — | User pressed Ctrl-C |
| `hilbish.cd` | `path, oldPath` | Directory changed |
| `hilbish.notification` | `notification` | Notification sent |
| `hilbish.init` | — | Shell initialization complete |

#### Job Events

| Event | Arguments | Description |
|-------|-----------|-------------|
| `job.add` | `job` | Job created |
| `job.start` | `job` | Job started |
| `job.done` | `job` | Job finished |

#### Signal Events

| Event | Arguments | Description |
|-------|-----------|-------------|
| `signal.sigint` | — | SIGINT received |
| `signal.resize` | — | Terminal resized |
| `signal.sigusr1` | — | SIGUSR1 |
| `signal.sigusr2` | — | SIGUSR2 |

### Usage Patterns

```lua
-- Update prompt on command exit
bait.catch('command.exit', function(code)
    doPrompt(code ~= 0)
end)

-- Custom not-found message
bait.catch('command.not-found', function(cmd)
    print(string.format('Command "%s" not found', cmd))
end)

-- Remove existing handler
local hooks = bait.hooks('command.not-found')
for _, hook in ipairs(hooks) do
    bait.release('command.not-found', hook)
end

-- One-time event
bait.catchOnce('hilbish.init', function()
    print('Hilbish is ready!')
end)
```

**Important**: Events are **global** — hook names must be unique across all modules. The `bait.throw()` function can be used to emit custom events.

## Runner Mode

Runner mode determines how hilbish interprets interactive input.

### Built-in Runners

| Runner | Description |
|--------|-------------|
| `hybrid` (default) | Try Lua first, fall back to shell |
| `hybridRev` | Try shell first, fall back to Lua |
| `lua` | Lua only |
| `sh` | Shell only |

### Runner API

```lua
hilbish.runner.add(name, runner)     -- Add a runner (function or table with run)
hilbish.runner.set(name, runner)     -- Set/overwrite a runner
hilbish.runner.get(name) → runner    -- Get a runner by name
hilbish.runner.setCurrent(name)      -- Set active runner
hilbish.runner.getCurrent() → string -- Get current runner name
hilbish.runner.exec(cmd, name?)      -- Execute cmd with runner
hilbish.runner.sh(input) → result    -- Run as shell script
hilbish.runner.lua(input) → result   -- Run as Lua code
```

### Custom Runners

A runner can be a function or a table with a `run` method:

```lua
-- Function runner
hilbish.runner.add('myRunner', function(input)
    -- process input
    return { exitCode = 0, input = input }
end)

-- Table runner
hilbish.runner.add('myRunner', {
    run = function(input)
        return { exitCode = 0, input = input }
    end
})

hilbish.runner.setCurrent('myRunner')
```

### Runner Return Value

The `run` function must return a table with:

| Field | Type | Description |
|-------|------|-------------|
| `exitCode` | number | Exit code of the execution |
| `input` | string | The input that was executed |
| `err` | string\|nil | Error message (triggers `command.not-found` or `command.not-executable`) |
| `continue` | boolean | Whether to prompt for more input |
| `newline` | boolean | Whether to prepend newline to continuation |

### Hybrid Runner Implementation

```lua
hilbish.runner.add('hybrid', function(input)
    local cmdStr = hilbish.aliases.resolve(input)
    local res = hilbish.runner.lua(cmdStr)
    if not res.err then
        return res
    end
    return hilbish.runner.sh(input)
end)
```

The `hybrid` runner resolves aliases before trying Lua, then falls back to shell if Lua fails.

## Commander Module

Commander allows creating Lua-implemented commands that appear as regular shell commands.

### API

| Function | Signature | Description |
|----------|-----------|-------------|
| `register` | `register(name, cb)` | Register a command |
| `deregister` | `deregister(name)` | Remove a command |
| `registry` | `registry() → table` | Get all registered commands |

### Command Callback

```lua
local commander = require 'commander'

commander.register('hello', function(args, sinks)
    local name = 'world'
    if #args > 0 then name = args[1] end
    sinks.out:writeln('Hello ' .. name)
end)
```

The callback receives:
- `args` — table of command arguments (1-indexed, excluding command name)
- `sinks` — table with `out`, `err`, and `input`/`in` (Sink objects)

### Sink Methods

| Method | Description |
|--------|-------------|
| `write(str)` | Write to sink |
| `writeln(str)` | Write with newline |
| `read() → string` | Read a line |
| `readAll() → string` | Read all content |
| `autoFlush(auto?)` | Toggle automatic flushing |
| `flush()` | Flush buffered output |

Commander commands appear in binary completion (via `cmds.Commands` in Go) but have no automatic argument completion.

## Lunacolors

Lunacolors is hilbish's ANSI color/styling library.

### Direct Functions

```lua
lunacolors.blue('text')
lunacolors.bold('text')
lunacolors.redBg('text')
```

### Format Function

```lua
lunacolors.format('{blue}%u {cyan}%d {green}∆ ')
```

Uses `{keyword}` placeholders. Automatically appends reset at the end.

### Available Colors

| Category | Names |
|----------|-------|
| Basic | `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white` |
| Bright | `brightBlack`, `brightRed`, `brightGreen`, `brightYellow`, `brightBlue`, `brightMagenta`, `brightCyan`, `brightWhite` |
| Background | `blackBg`, `redBg`, `greenBg`, `yellowBg`, `blueBg`, `magentaBg`, `cyanBg`, `whiteBg` |
| Bright BG | `brightBlackBg`, `brightRedBg`, `brightGreenBg`, `brightYellowBg`, `brightBlueBg`, `brightMagentaBg`, `brightCyanBg`, `brightWhiteBg` |

### Available Styles

`reset`, `bold`, `dim`, `italic`, `underline`, `invert`, `hidden`, `strikethrough`

### Format Syntax (in `lunacolors.format`)

In the format function, keywords use hyphens:
- `{blue}`, `{red-bg}`, `{bright-yellow}`
- `{bold}`, `{italic}`, `{reset}`

## Yarn (Multi-threading)

Yarn provides multi-threading via separate Lua states.

### API

```lua
local yarn = require 'yarn'
local t = yarn.thread(function(msg)
    print("Thread received:", msg)
end)
t "Hello from another thread!"
```

### Thread Isolation

- Each thread has its **own Lua state** — no shared environment
- **Bait hooks** are shared between threads (inter-thread communication)
- **Commanders** are shared between threads
- Threads run in **goroutines** — non-blocking

**Warning**: Accessing variables from the main Lua state in a yarn thread may crash hilbish (GopherLua limitation).

## The `fs` Module

| Function | Signature | Description |
|----------|-----------|-------------|
| `abs` | `abs(path) → string` | Absolute path |
| `basename` | `basename(path) → string` | Last component of path |
| `cd` | `cd(dir)` | Change directory |
| `dir` | `dir(path) → string` | Directory part of path |
| `glob` | `glob(pattern) → table` | Glob match (Go `filepath.Match` syntax) |
| `join` | `join(...path) → string` | Join paths with OS separator |
| `mkdir` | `mkdir(name, recursive)` | Create directory |
| `fpipe` | `fpipe() → File, File` | Create pipe (read end, write end) |
| `readdir` | `readdir(path) → table` | List directory entries |
| `stat` | `stat(path) → table` | File info: name, size, mode, isDir |

Module field: `fs.pathSep` — OS path separator.

## The `terminal` Module

Low-level terminal operations:

| Function | Description |
|----------|-------------|
| `restoreState()` | Restore saved terminal state |
| `saveState()` | Save current terminal state |
| `setRaw()` | Put terminal in raw mode |
| `size() → {width, height}` | Get terminal dimensions |

## Quoting and Expansion

Hilbish's shell interpreter (via the `snail` module and Go's `mvdan.cc/sh/v3`) handles quoting and expansion for the `sh` and `hybrid` runner modes.

### Shell Quoting (sh mode)

- **Single quotes** — literal, no expansion: `'hello world'`
- **Double quotes** — variable expansion: `"hello $USER"`
- **Backslash** — escape next character: `hello\ world`
- **Tilde** — home directory expansion: `~/Documents`

### Lua Quoting (lua mode)

Standard Lua quoting rules apply:
- **Single quotes** — `'hello'`
- **Double quotes** — `"hello"` (with escape sequences)
- **Long strings** — `[[hello]]` or `[=[hello]=]`

### Character Escaping in Completions

The Go completion layer escapes these characters in file completion results:

```
" ' ` space ( ) [ ] $ & * > < |
```

This is handled by `escapeFilename()` using `strings.NewReplacer(charEscapeMap...)`.

## Edge Cases and Known Issues

- **`hilbish.goro()` is dangerous**: Accessing variables from the main Lua state in a goroutine can crash hilbish. This is a fundamental GopherLua limitation — the runtime is not thread-safe.
- **`hilbish.runnerMode` is deprecated**: Use `hilbish.runner.setCurrent()` instead. Will be removed in 3.0.
- **No `table.filter` in standard Lua**: The `sudo.lua` completion uses `table.filter()`, which is provided by the `succulent` library (loaded in `nature/init.lua` via `require 'succulent'`).
- **`string.split` is not standard Lua**: Hilbish adds `string.split` as an extension.
- **Commander commands lack completion**: Lua-registered commands appear in binary completion but have no argument completion unless explicitly registered via `hilbish.completions.add()`.

## References

- Hilbish API: <https://hilbish.sammyette.party/docs/api/hilbish>
- Bait API: <https://hilbish.sammyette.party/docs/api/bait>
- Commander API: <https://hilbish.sammyette.party/docs/api/commander>
- Runner mode: <https://hilbish.sammyette.party/docs/features/runner-mode>
- Lunacolors: <https://hilbish.sammyette.party/docs/lunacolors>
- Yarn API: <https://hilbish.sammyette.party/docs/api/yarn>
- FS API: <https://hilbish.sammyette.party/docs/api/fs>
- Terminal API: <https://hilbish.sammyette.party/docs/api/terminal>
- Source: `api.go` — Core hilbish module
- Source: `nature/hilbish.lua` — `hilbish.run()`, `hilbish.runnerMode()`
- Source: `nature/runner.lua` — Runner mode system
- Source: `nature/init.lua` — Bootstrap and module loading

## Related Skills

- [Completion](completion.md) — How completion handlers are registered and called
- [Line Editing](line-editing.md) — How the readline integrates with Lua callbacks
- [Execution](execution.md) — How commands are executed and jobs managed
- [Startup & Configuration](startup-config.md) — How hilbish initializes and loads modules
