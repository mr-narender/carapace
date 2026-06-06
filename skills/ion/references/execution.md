# Ion Shell Execution Model

In-depth reference for ion's execution model — job control, signals, process management, and exit status handling.

## Command Execution Flow

When ion executes a command:

1. **Parse** — the input line is tokenized and parsed into an AST
2. **Expand** — variables, brace expansion, command substitution, glob expansion
3. **Resolve** — locate the command (builtin, alias, function, or external)
4. **Execute** — run the command in the appropriate context

### Command Resolution Order

1. **Built-in commands** — `cd`, `echo`, `test`, `let`, etc.
2. **Aliases** — expanded and re-resolved
3. **Functions** — user-defined with `fn`
4. **External commands** — searched in `$PATH`

## Pipelines

### stdout Pipeline

```sh
cmd1 | cmd2
```

- `cmd1`'s stdout is connected to `cmd2`'s stdin via a pipe
- All commands in a pipeline run concurrently as separate processes
- The exit status of the pipeline is the exit status of the **last command**

### stderr Pipeline (Ion Unique)

```sh
cmd1 ^| cmd2
```

- `cmd1`'s **stderr** is piped to `cmd2`'s stdin
- stdout is not redirected
- No other shell provides this as a first-class operator

### Combined Pipeline (Ion Unique)

```sh
cmd1 &| cmd2
```

- Both stdout and stderr are piped to `cmd2`'s stdin
- Equivalent to `cmd1 2>&1 | cmd2` in POSIX shells

## Logical Operators

### AND (`&&`)

```sh
cmd1 && cmd2
```

- `cmd2` runs only if `cmd1` exits with status 0
- Short-circuit: if `cmd1` fails, `cmd2` is not executed

### OR (`||`)

```sh
cmd1 || cmd2
```

- `cmd2` runs only if `cmd1` exits with non-zero status
- Short-circuit: if `cmd1` succeeds, `cmd2` is not executed

### Combining

```sh
cmd1 && cmd2 || cmd3
# If cmd1 succeeds, run cmd2; if cmd2 fails, run cmd3
```

## Background Jobs

### Running in Background

```sh
command &
```

- The command runs in the background
- The shell does not wait for it to complete
- The shell prompt returns immediately

### Job Control Commands

| Command | Description |
|---------|-------------|
| `jobs` | List all background jobs |
| `fg PID` | Bring a job to the foreground |
| `bg PID` | Resume a stopped job in the background |
| `disown PID` | Remove a job from the shell's job table |
| `disown -a` | Disown all jobs |
| `disown -h PID` | Disown but keep running (no SIGHUP on exit) |
| `wait` | Wait for all background jobs to complete |
| `suspend` | Suspend the shell (send SIGTSTP to self) |

### Job Lifecycle

```
[Running] → Ctrl+Z → [Stopped]
[Stopped] → bg → [Running (background)]
[Stopped] → fg → [Running (foreground)]
[Running] → disown → [Detached from shell]
```

## Signals

### Signal Handling

Ion handles signals as follows:

| Signal | Behavior |
|--------|----------|
| `SIGINT` (Ctrl+C) | Interrupt current command; in interactive mode, cancel current line and show prompt |
| `SIGTSTP` (Ctrl+Z) | Suspend current foreground job |
| `SIGHUP` | Sent to all jobs when the shell exits (unless disowned with `-h`) |
| `SIGTERM` | Terminate the shell |
| `SIGQUIT` (Ctrl+\\) | Quit with core dump |

### SIGHUP and Shell Exit

When an interactive ion shell exits:

1. SIGHUP is sent to all jobs in the job table
2. Jobs disowned with `disown -h` are **not** sent SIGHUP
3. Jobs disowned with `disown` (without `-h`) are removed from the table entirely

### Exit Status

- Exit code `0` = success (true)
- Exit code non-zero = failure (false)
- The `$?` variable is **not** available in ion — use `and`/`or`/`not` for status checking

## Process Replacement

### exec

```sh
exec command args...
```

- Replaces the current shell process with the given command
- No subshell is created — the shell process itself is replaced
- Options:
  - `exec -c` — clear environment before executing
  - `exec -h` — show help

### eval

```sh
eval command args...
```

- Concatenates arguments and executes as a command
- Useful for constructing commands dynamically

### source

```sh
source file
```

- Execute commands from `file` in the current shell context
- Variables and functions defined in the file persist

## Subshells

Ion does not have a traditional subshell syntax like `( commands )` in bash. Instead:

- **Command substitution** creates a subprocess: `$(command)`
- **Pipelines** run each command in a separate process
- **Background jobs** run in separate processes

## Exit Status and Conditionals

### test Builtin

```sh
test $x -eq 5
test -f /path/to/file
test "hello" = "hello"
```

Returns exit status 0 (true) or 1 (false).

### and / or / not

```sh
and cmd1 cmd2    # cmd2 only if cmd1 succeeds
or cmd1 cmd2     # cmd2 only if cmd1 fails
not cmd          # invert exit status
```

### Conditional Execution in If

```sh
if test -f file
    echo "exists"
else
    echo "not found"
end
```

## Redirection and File Descriptors

### Standard File Descriptors

| FD | Name | Default |
|----|------|---------|
| 0 | stdin | Terminal input |
| 1 | stdout | Terminal output |
| 2 | stderr | Terminal error output |

### Redirection Operators

| Operator | Source | Destination | Example |
|----------|--------|-------------|---------|
| `>` | stdout | File (truncate) | `cmd > out.txt` |
| `>>` | stdout | File (append) | `cmd >> out.txt` |
| `<` | File | stdin | `cmd < in.txt` |
| `^>` | stderr | File (truncate) | `cmd ^> err.txt` |
| `^>>` | stderr | File (append) | `cmd ^>> err.txt` |
| `&>` | stdout+stderr | File (truncate) | `cmd &> all.txt` |
| `&>>` | stdout+stderr | File (append) | `cmd &>> all.txt` |
| `<<` | stdin | Here-document | `cmd << 'EOF'` |
| `<<<` | String | stdin | `cmd <<< "text"` |

### Pipe Operators

| Operator | Source | Destination | Example |
|----------|--------|-------------|---------|
| `\|` | stdout | stdin of next | `cmd1 \| cmd2` |
| `^\|` | stderr | stdin of next | `cmd1 ^\| cmd2` |
| `&\|` | stdout+stderr | stdin of next | `cmd1 &\| cmd2` |

## Error Handling

### set -e

```sh
set -e
```

Exit immediately if any command exits with a non-zero status. This is useful for scripts that should fail fast.

### set -x

```sh
set -x
```

Print each command before executing it (trace mode). Useful for debugging scripts.

## Comparison with POSIX Shells

| Feature | Ion | Bash | POSIX |
|---------|-----|------|-------|
| stderr pipe | `^\|` (native) | `2>&1 \|` (redirect) | `2>&1 \|` (redirect) |
| combined pipe | `&\|` (native) | `2>&1 \|` (redirect) | `2>&1 \|` (redirect) |
| stderr redirect | `^>` (native) | `2>` (FD syntax) | `2>` (FD syntax) |
| Subshell syntax | None | `( cmds )` | `( cmds )` |
| `$?` variable | Not available | Available | Available |
| `set -e` | Supported | Supported | Supported |
| `set -x` | Supported | Supported | Supported |
| Job control | `bg`/`fg`/`jobs`/`disown` | Same + `wait PID` | Same |

## References

- [Ion Manual — Control Flow](https://doc.redox-os.org/ion-manual/control/00-flow.html)
- [Ion Manual — Builtins](https://doc.redox-os.org/ion-manual/builtins.html)
- [Ion Manual — Scripts](https://doc.redox-os.org/ion-manual/scripts.html)
- [Ion GitHub Repository](https://github.com/redox-os/ion)

## Related Skills

- [references/language.md](references/language.md) — ion syntax including redirection operators
- [references/startup-config.md](references/startup-config.md) — shell options and configuration
- [references/completion.md](references/completion.md) — completion system
