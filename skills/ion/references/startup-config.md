# Ion Shell Startup and Configuration

In-depth reference for ion's startup files, configuration, keybindings, prompt customization, history, and plugins.

## Startup File Decision Flow

### Interactive Shell

Ion reads a single startup file:

1. **`~/.config/ion/initrc`** — the ion initialization script

If `$XDG_CONFIG_HOME` is set, the path becomes `$XDG_CONFIG_HOME/ion/initrc`.

### Non-Interactive Shell (Script Execution)

Ion does **not** read `initrc` when running scripts. There is no equivalent to bash's `BASH_ENV`.

### Script Execution

```sh
#!/usr/bin/env ion

echo "Arguments: @args[1..]"
```

The shebang `#!/usr/bin/env ion` is the standard way to write ion scripts.

## The initrc File

### Location

| Condition | Path |
|-----------|------|
| Default | `~/.config/ion/initrc` |
| `$XDG_CONFIG_HOME` set | `$XDG_CONFIG_HOME/ion/initrc` |

### Example initrc

```sh
# Keybindings
keybindings vi

# Aliases
alias ll = "ls -la"
alias gs = "git status"
alias gp = "git push"

# Environment variables
export EDITOR = vim
export PATH = $PATH:~/bin

# Prompt
fn PROMPT
    echo -n "${color::green}${env::USER}${color::reset}@"
    echo -n "${color::cyan}${env::HOST}${color::reset}:"
    echo -n "${color::blue}$(pwd)${color::reset}$ "
end

# History settings
history +inc_append
history -duplicates
```

### What Can Go in initrc

| Construct | Example |
|-----------|---------|
| Keybinding mode | `keybindings vi` |
| Aliases | `alias ll = "ls -la"` |
| Environment variables | `export EDITOR = vim` |
| Variable declarations | `let default_branch = "main"` |
| Function definitions | `fn PROMPT ... end` |
| History configuration | `history +inc_append` |
| Source other files | `source ~/.config/ion/aliases` |
| Shell options | `set -o vi` |

## Keybindings

### Switching Modes

```sh
keybindings emacs    # default mode
keybindings vi       # vi mode

# Alternative syntax
set -o emacs
set -o vi
```

### Emacs Mode (Default)

| Key | Action |
|-----|--------|
| `Ctrl+A` | Move to beginning of line |
| `Ctrl+E` | Move to end of line |
| `Ctrl+B` | Move left one character |
| `Ctrl+F` | Move right one character / Accept autosuggestion |
| `Ctrl+P` | Previous history entry |
| `Ctrl+N` | Next history entry |
| `Ctrl+D` | Delete character after cursor |
| `Ctrl+K` | Kill to end of line |
| `Ctrl+U` | Kill to beginning of line |
| `Ctrl+W` | Kill previous word |
| `Ctrl+L` | Clear screen |
| `Ctrl+R` | Reverse incremental search |
| `Ctrl+X` | Undo |
| `Alt+B` | Move backward one word |
| `Alt+F` | Move forward one word |
| `Alt+<` | Move to first history entry |
| `Alt+>` | Move to last history entry |
| `Alt+.` | Insert last argument of previous command |
| `Alt+R` | Revert line to original |
| `Tab` | Complete |
| `Right Arrow` (at EOL) | Accept autosuggestion |

### Vi Mode

| Mode | Key | Action |
|------|-----|--------|
| Normal | `i` | Enter insert mode |
| Normal | `a` | Append after cursor |
| Normal | `A` | Append at end of line |
| Normal | `I` | Insert at beginning of line |
| Normal | `h` / `l` | Move left / right |
| Normal | `j` / `k` | Next / previous history |
| Normal | `w` / `b` | Forward / backward word |
| Normal | `0` / `$` | Beginning / end of line |
| Normal | `f` + char | Find character forward |
| Normal | `F` + char | Find character backward |
| Normal | `gg` | First history entry |
| Normal | `G` | Last history entry |
| Normal | `dd` | Delete line |
| Normal | `dw` | Delete word |
| Normal | `cc` | Change line |
| Normal | `cw` | Change word |
| Normal | `x` | Delete character |
| Normal | `r` | Replace character |
| Normal | `.` | Repeat last command |
| Normal | `/` | Search forward |
| Normal | `v` | Visual mode |
| Insert | `Esc` | Return to normal mode |
| Insert | `Tab` | Complete |

### Vi Mode Indicators

```sh
export VI_NORMAL = "[=] "
export VI_INSERT = "[+] "
keybindings vi
```

## Prompt Customization

### The PROMPT Function

Ion uses a function named `PROMPT` to generate the prompt string:

```sh
fn PROMPT
    echo -n "${color::green}${env::USER}${color::reset}@"
    echo -n "${color::cyan}${env::HOST}${color::reset}:"
    echo -n "${color::blue}$(pwd)${color::reset}$ "
end
```

**Key details:**
- The function is called every time the prompt is displayed
- Use `echo -n` to avoid trailing newline
- Colors are applied via the `${color::name}` namespace
- Command substitution (`$(pwd)`) is evaluated each time

### Color Namespace

Available color names:

| Category | Names |
|----------|-------|
| Basic | `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white` |
| Bright | `light_black`, `light_red`, `light_green`, `light_yellow`, `light_blue`, `light_magenta`, `light_cyan`, `light_white` |
| Special | `reset` (reset all attributes) |

**Usage:**

```sh
${color::green}     # foreground green
${color::reset}     # reset to default
```

### Environment Namespace

```sh
${env::USER}        # $USER environment variable
${env::HOME}        # $HOME environment variable
${env::HOST}        # hostname
```

### Third-Party Prompt Tools

| Tool | Description |
|------|-------------|
| [Starship](https://starship.rs/) | Cross-shell prompt (supports ion) |
| [candypaint](https://gitlab.redox-os.org/redox-os/candypaint) | Zero-config prompts for ion |
| [grainion](https://gitlab.redox-os.org/redox-os/grainion) | Lambda prompt for RedoxOS |

## History Configuration

### The history Builtin

```sh
history [options]
```

| Option | Description |
|--------|-------------|
| `+inc_append` | Append each command to history immediately |
| `-inc_append` | Don't append until exit (default) |
| `+shared` | Share history between shells (implies inc_append) |
| `-shared` | Don't share (default) |
| `+duplicates` | Allow duplicate entries (default) |
| `-duplicates` | Don't allow duplicate entries |

### History File

Default location: `~/.local/share/ion/history`

Override with:

```sh
export HISTFILE = ~/.custom/history
```

### HISTORY_IGNORE

```sh
export HISTORY_IGNORE = [
    'no_such_command'     # exact match
    'whitespace'          # commands starting with space
    'duplicates'          # duplicate entries
    regex:'^cd '          # regex pattern
]
```

### Autosuggestions

Ion supports fish-like autosuggestions from history:

- Shows matching history entries in **yellow** as you type
- Accept with `Ctrl+F` or `Right Arrow` at end of line
- Only the newest matching history entry is shown
- Controlled by `show_autosuggestions` in the Liner Context

## Shell Options

### set Builtin

```sh
set -e          # Exit on error
set -x          # Trace mode (print commands before execution)
set -o vi       # Vi keybindings
set -o emacs     # Emacs keybindings
```

## Aliases

### Defining Aliases

```sh
alias ll = "ls -la"
alias gs = "git status"
```

### Removing Aliases

```sh
unalias ll
```

### Listing Aliases

```sh
alias           # list all aliases
```

## ion-plugins

### Overview

The [ion-plugins](https://gitlab.redox-os.org/redox-os/ion-plugins) repository provides additional aliases and function definitions written in ion for ion.

### What Plugins Can Do

| Feature | Supported |
|---------|-----------|
| Aliases | Yes |
| Functions | Yes |
| Environment variables | Yes |
| Custom completions | **No** |
| Key bindings | No (set in initrc) |

### Installing Plugins

```sh
# Clone the plugins repository
git clone https://gitlab.redox-os.org/redox-os/ion-plugins.git ~/.config/ion/plugins

# Source in initrc
source ~/.config/ion/plugins/aliases.ion
source ~/.config/ion/plugins/functions.ion
```

### Limitations

- Plugins **cannot** add custom completion logic
- Plugins **cannot** modify key bindings
- Plugins are just ion scripts — they define aliases and functions
- There is no package manager for ion plugins

## XDG Directory Layout

```
~/.config/ion/
  initrc              # Startup script

~/.local/share/ion/
  history             # History file (default)
```

Override with environment variables:

| Variable | Default | Purpose |
|----------|---------|---------|
| `XDG_CONFIG_HOME` | `~/.config` | Configuration directory |
| `XDG_DATA_HOME` | `~/.local/share` | Data directory |
| `HISTFILE` | `$XDG_DATA_HOME/ion/history` | History file path |

## Comparison with Other Shells

| Feature | Ion | Bash | Fish | Zsh |
|---------|-----|------|------|-----|
| Startup file | `initrc` (single) | `.bashrc` + `.bash_profile` + `.profile` | `config.fish` | `.zshrc` + `.zprofile` + ... |
| XDG support | Native | Via env vars | Native | Via env vars |
| Keybinding modes | vi, emacs | vi, emacs | vi, emacs, fish | vi, emacs |
| Prompt function | `fn PROMPT` | `PS1` variable | `fish_prompt` function | `PROMPT`/`PS1` |
| Color in prompt | `${color::name}` | ANSI escapes | `set_color` | `%F{color}` |
| History sharing | `history +shared` | `shopt -s histappend` | Built-in | `SHARE_HISTORY` |
| Plugin system | ion-plugins (scripts only) | bash-completion + plugins | Fisher/Oh My Fish | Oh My Zsh/zinit |
| Custom completions in plugins | **No** | Yes | Yes | Yes |

## References

- [Ion Manual — Scripts](https://doc.redox-os.org/ion-manual/scripts.html)
- [Ion Manual — Builtins](https://doc.redox-os.org/ion-manual/builtins.html)
- [ion-plugins Repository](https://gitlab.redox-os.org/redox-os/ion-plugins)
- [Starship Prompt](https://starship.rs/)
- [Ion GitHub Repository](https://github.com/redox-os/ion)

## Related Skills

- [references/line-editing.md](references/line-editing.md) — the Liner library that handles key bindings and editing
- [references/completion.md](references/completion.md) — completion system configuration
- [references/language.md](references/language.md) — ion syntax for initrc content
- [references/execution.md](references/execution.md) — shell options (`set -e`, `set -x`)
