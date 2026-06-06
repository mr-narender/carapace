# Hilbish Execution Model

In-depth reference for hilbish's command execution, job control, signal handling, and the runner dispatch pipeline.

## Overview

Hilbish executes commands through a **runner dispatch** system. The active runner (hybrid, sh, lua, or custom) determines how input is interpreted. Shell commands are executed via the `snail` module (wrapping `mvdan.cc/sh/v3`), while Lua code runs through the GopherLua runtime. Background jobs are managed by a thread-safe `jobHandler` with mutex-protected state.

## Source Files

| File | Language | Purpose |
|------|----------|---------|
| `exec.go` | Go | `runInput()`, `handleLua()`, `splitInput()` |
| `job.go` | Go | `job` struct, `jobHandler`, all job management |
| `nature/runner.lua` | Lua | Runner mode system, `hilbish.runner.run()` |
| `nature/hilbish.lua` | Lua | `hilbish.run()`, `hilbish.runnerMode()` |
| `main.go` | Go | Main loop, `continuePrompt()`, `exit()` |

## Command Execution Pipeline

### Interactive Input Flow

```
User types input + Enter
  → lr.Read()                    # readline reads line
  → runInput(input, priv)        # Go entry point
  → hilbish.runner.run(input)    # Lua runner dispatch
  → bait.throw('command.preprocess', input)
  → hilbish.processors.execute() # command processors
  → hilbish.aliases.resolve()    # alias expansion
  → bait.throw('command.preexec', input, command)
  → runner.run(command)          # active runner executes
  → finishExec(exitCode, input)  # cleanup
  → bait.throw('command.exit', exitCode, input)
```

### `runInput()` (Go)

```go
func runInput(input string, priv bool) {
    running = true
    runnerRun := hshMod.Get(rt.StringValue("runner")).AsTable().Get(rt.StringValue("run"))
    _, err := rt.Call1(l.MainThread(), runnerRun, rt.StringValue(input), rt.BoolValue(priv))
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}
```

This calls the Lua `hilbish.runner.run()` function, which is the main execution dispatch.

### `hilbish.runner.run()` (Lua)

```lua
function hilbish.runner.run(input, priv)
    bait.throw('command.preprocess', input)
    local processed = hilbish.processors.execute(input, {
        skip = hilbish.opts.processorSkipList
    })
    priv = processed.history ~= nil and (not processed.history) or priv
    if not processed.continue then
        finishExec(0, '', true)
        return
    end

    local command = hilbish.aliases.resolve(processed.command)
    bait.throw('command.preexec', processed.command, command)

    ::rerun::
    local runner = hilbish.runner.get(currentRunner)
    local ok, out = pcall(runner.run, processed.command)
    if not ok then
        io.stderr:write(out .. '\n')
        finishExec(124, out.input, priv)
        return
    end

    if out.continue then
        local contInput = continuePrompt(processed.command, out.newline)
        if contInput then
            processed.command = contInput
            goto rerun
        end
    end

    if out.err then
        local fields = string.split(out.err, ': ')
        if fields[2] == 'not-found' or fields[2] == 'not-executable' then
            bait.throw('command.' .. fields[2], fields[1])
        else
            io.stderr:write(out.err .. '\n')
        end
    end
    finishExec(out.exitCode, out.input, priv)
end
```

Key steps:

1. **Preprocess** — emits `command.preprocess` event
2. **Processors** — command processors (e.g., wildcard warning) can modify or skip execution
3. **Alias resolution** — `hilbish.aliases.resolve()` expands aliases
4. **Pre-exec hook** — emits `command.preexec` with original and resolved command
5. **Runner dispatch** — calls the active runner's `run()` function
6. **Continuation** — if `out.continue` is true, prompts for more input
7. **Error handling** — maps `not-found` and `not-executable` errors to bait events
8. **Finish** — sets exit code and emits `command.exit`

### Private Mode

If input starts with a space, `priv` is `true` and the command is not saved to history:

```go
var priv bool
if strings.HasPrefix(input, " ") {
    priv = true
}
```

## Shell Execution (sh runner)

The `sh` runner uses the `snail` module, which wraps `mvdan.cc/sh/v3`:

```lua
function hilbish.runner.sh(input)
    return hilbish.snail:run(input)
end
```

The `snail` module provides a POSIX-compatible shell interpreter. It handles:
- Command execution via `os/exec`
- Pipelines and redirections
- Variable expansion and globbing
- Subshells and command substitution

### `hilbish.run()` — Programmatic Shell Execution

```lua
function hilbish.run(cmd, streams)
    local sinks = {}
    if type(streams) == 'boolean' then
        if not streams then
            sinks = {
                out = hilbish.sink.new(),
                err = hilbish.sink.new(),
                input = io.stdin
            }
        end
    elseif type(streams) == 'table' then
        sinks = streams
    end

    local out = hilbish.snail:run(cmd, {sinks = sinks})
    local returns = {out.exitCode}

    if type(streams) == 'boolean' and not streams then
        table.insert(returns, sinks.out:readAll())
        table.insert(returns, sinks.err:readAll())
    end

    return table.unpack(returns)
end
```

## Lua Execution (lua runner)

The `lua` runner compiles and executes Lua code:

```go
func handleLua(input string) (string, uint8, error) {
    cmdString := aliases.Resolve(input)
    chunk, err := l.CompileAndLoadLuaChunk("", []byte(cmdString), rt.TableValue(l.GlobalEnv()))
    if err != nil && noexecute {
        return cmdString, 125, err
    }
    if !noexecute {
        if chunk != nil {
            _, err = rt.Call1(l.MainThread(), rt.FunctionValue(chunk))
        }
    }
    if err == nil {
        return cmdString, 0, nil
    }
    return cmdString, 125, err
}
```

Key details:
- Aliases are resolved before Lua execution
- The `--noexec`/`-n` flag compiles but does not execute (syntax check only)
- Lua errors return exit code 125

## Job Control

### The `job` Struct

```go
type job struct {
    cmd       string          // User-entered command string
    running   bool            // Whether the job is running
    id        int             // Job ID in the table
    pid       int             // Process ID
    exitCode  int             // Last exit code
    once      bool            // Whether the job has been started once
    args      []string        // Command arguments
    path      string          // Absolute path to binary
    handle    *exec.Cmd       // Go command handle
    cmdout    io.Writer       // Standard output destination
    cmderr    io.Writer       // Standard error destination
    stdout    *bytes.Buffer   // Output buffer for Lua access
    stderr    *bytes.Buffer   // Error buffer for Lua access
    ud        *rt.UserData     // Lua userdata wrapper
}
```

### Job Lifecycle

```
jobs.add(cmd, args, path)  →  job created, 'job.add' hook emitted
  → job.start()            →  process started, 'job.start' hook emitted
  → job.wait()             →  blocks until process exits
  → job.finish()           →  'job.done' hook emitted, running = false
```

Or for stopped jobs:
```
  → job.stop()             →  process killed
  → job.finish()           →  'job.done' hook emitted
```

### `jobHandler`

```go
type jobHandler struct {
    jobs       map[int]*job
    latestID   int
    foreground bool          // Whether a job is currently in the foreground
    mu         *sync.RWMutex // Thread-safe access
}
```

All job operations are mutex-protected for concurrent access from goroutines and Lua threads.

### Job Lua API

| Function | Signature | Description |
|----------|-----------|-------------|
| `all` | `all() → table` | Get all jobs |
| `last` | `last() → Job` | Get most recently added job |
| `get` | `get(id) → Job` | Get job by ID |
| `add` | `add(cmdstr, args, execPath) → Job` | Create a new job (does not start it) |
| `disown` | `disown(id)` | Remove job from table without stopping |

### Job Object Methods

| Method | Description |
|--------|-------------|
| `start()` | Start the job |
| `stop()` | Kill the job process |
| `foreground()` | Bring job to foreground (waits for completion) |
| `background()` | Send job to background (continues running) |

### Job Object Properties

| Property | Type | Description |
|----------|------|-------------|
| `cmd` | string | Command string |
| `running` | boolean | Whether running |
| `id` | number | Job ID |
| `pid` | number | Process ID |
| `exitCode` | number | Exit code |
| `stdout` | string | Captured standard output |
| `stderr` | string | Captured standard error |

### Foreground/Background

When a job is brought to the foreground:

```go
func luaForegroundJob(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
    jobs.foreground = true
    err = j.background()  // resume if suspended
    err = j.foreground()  // bring to foreground
    jobs.foreground = false
}
```

The `foreground` flag on `jobHandler` prevents signal forwarding to background jobs while a foreground job is active.

### Process Attributes

Background jobs use `bgProcAttr` (defined in `job_<os>.go`) to prevent signals from hilbish (like SIGINT) from being forwarded to the child process:

```go
j.handle.SysProcAttr = bgProcAttr
```

## Signal Handling

### Go Signal Handler

Hilbish runs a goroutine to handle OS signals:

```go
go handleSignals()
```

### Signal Events

| Signal | Bait Event | Description |
|--------|-----------|-------------|
| SIGINT | `signal.sigint` | Ctrl-C pressed |
| SIGTERM | (exit) | Terminal closed |
| SIGHUP | (sent to all jobs on exit) | Shell exiting |
| SIGUSR1 | `signal.sigusr1` | User-defined |
| SIGUSR2 | `signal.sigusr2` | User-defined |
| SIGWINCH | `signal.resize` | Terminal resized |

### Exit Cleanup

On exit, hilbish:

1. Sends SIGHUP to all running jobs via `jobs.stopAll()`
2. Waits for each job to exit
3. If non-interactive, waits for all timers to finish
4. Calls `os.Exit(code)`

```go
func exit(code int) {
    jobs.stopAll()
    if !interactive {
        timers.wait()
    }
    os.Exit(code)
}
```

### Ctrl-C in Interactive Mode

When Ctrl-C is pressed during input:

```go
if err == readline.CtrlC {
    fmt.Println("^C")
    hooks.Emit("hilbish.cancel")
}
```

The `hilbish.cancel` event is emitted, allowing Lua code to clean up state.

## Command Not Found / Not Executable

When a command fails with specific error types, hilbish emits targeted events:

```lua
-- In hilbish.runner.run():
if out.err then
    local fields = string.split(out.err, ': ')
    if fields[2] == 'not-found' or fields[2] == 'not-executable' then
        bait.throw('command.' .. fields[2], fields[1])
    else
        io.stderr:write(out.err .. '\n')
    end
end
```

The default error handlers (from `nature/init.lua`):

```lua
bait.catch('command.not-found', function(cmd)
    print(string.format('hilbish: %s not found', cmd))
end)

bait.catch('command.not-executable', function(cmd)
    print(string.format('hilbish: %s: not executable', cmd))
end)
```

## Multiline Input

When input ends with `\`, hilbish enters continuation mode:

```go
if strings.HasSuffix(input, "\\") {
    print("\n")
    for {
        input, err = continuePrompt(strings.TrimSuffix(input, "\\") + "\n", false)
        if err != nil { goto input }
        if !strings.HasSuffix(input, "\\") { break }
    }
}
```

The `continuePrompt()` function:

```go
func continuePrompt(prev string, newline bool) (string, error) {
    hooks.Emit("multiline", nil)
    lr.SetPrompt(multilinePrompt)
    cont, err := lr.Read()
    if newline { cont = "\n" + cont }
    if strings.HasSuffix(cont, "\\") {
        cont = strings.TrimSuffix(cont, "\\") + "\n"
    }
    return prev + cont, nil
}
```

## Edge Cases and Known Issues

- **Exit code 124 for runner errors**: If the runner itself fails (not the command), exit code 124 is used. This mirrors bash's `command not found` exit code convention.
- **Exit code 125 for Lua errors**: Lua compilation/execution errors return exit code 125.
- **`jobs.foreground` flag**: This is a simple boolean, not a queue. Only one job can be in the foreground at a time.
- **SIGHUP on exit**: All running jobs receive SIGHUP when hilbish exits, matching POSIX shell behavior.
- **No job control builtins**: Unlike bash's `fg`, `bg`, `jobs`, `disown` builtins, hilbish exposes job management only through the Lua API.
- **`splitInput()` is simplistic**: The Go `splitInput()` function only handles double quotes and tilde expansion. It does not handle single quotes, escaped quotes, or complex shell syntax.

## References

- Hilbish hooks documentation: <https://hilbish.sammyette.party/docs/hooks>
- Source: `exec.go` — `runInput()`, `handleLua()`, `splitInput()`
- Source: `job.go` — `job` struct, `jobHandler`, all job methods
- Source: `nature/runner.lua` — Runner dispatch and `hilbish.runner.run()`
- Source: `nature/hilbish.lua` — `hilbish.run()`
- Source: `main.go` — Main loop, `continuePrompt()`, `exit()`

## Related Skills

- [Language & API](language.md) — Runner mode system, bait events, and the Lua API
- [Line Editing](line-editing.md) — How readline handles input and continuation prompts
- [Startup & Configuration](startup-config.md) — How hilbish initializes and configures runners
