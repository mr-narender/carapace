# Hilbish Line Editing

In-depth reference for hilbish's line editing system — the `maxlandon/readline` library, the `lineReader` Go wrapper, the `hilbish.editor` Lua API, vim mode, syntax highlighting, hints, and history search.

## Overview

Hilbish uses a forked version of the `maxlandon/readline` Go library for its line editor. Unlike bash (GNU Readline) or zsh (ZLE), hilbish's readline is implemented entirely in Go with Lua callbacks for completion, highlighting, and hints. The line editor is created in `rl.go` via `newLineReader()` and exposed to Lua as `hilbish.editor`.

## Source Files

| File | Language | Purpose |
|------|----------|---------|
| `rl.go` | Go | `lineReader` struct, `newLineReader()`, all readline callbacks, history Lua module |
| `nature/editor.lua` | Lua | `hilbish.editor` wrapper around `readline.new()` |
| `nature/vim.lua` | Lua | `hilbish.vim` registers and mode access |

## The `lineReader` Struct

```go
type lineReader struct {
    rl       *readline.Readline
    fileHist *fileHistory
}
```

The `lineReader` wraps a `readline.Readline` instance and a file-based history store.

## `newLineReader()` — Initialization

Creates and configures the readline instance with all custom callbacks:

```go
func newLineReader(prompt string, noHist bool) *lineReader {
    rl := readline.NewInstance()
    lr := &lineReader{rl: rl}

    // Configure: fuzzy search, vim mode, hints, highlighting, tab completion
    // ...
    return lr
}
```

### Callbacks Registered

| Callback | Go Field | Lua Integration | Purpose |
|----------|----------|----------------|---------|
| `rl.Searcher` | Custom | `hilbish.opts.fuzzy` | Fuzzy or regex history search |
| `rl.ViModeCallback` | Custom | `setVimMode()` + `hilbish.vimMode` hook | Vim mode change notification |
| `rl.ViActionCallback` | Custom | `hilbish.vimAction` hook | Vim yank/paste action notification |
| `rl.HintText` | Custom | `hilbish.hinter(line, pos)` | Inline hint text after cursor |
| `rl.SyntaxHighlighter` | Custom | `hilbish.highlighter(line)` | Syntax highlighting of input line |
| `rl.TabCompleter` | Custom | `hilbish.completions.handler(line, pos)` | Tab completion dispatch |

## Fuzzy History Search

The `Searcher` callback replaces the default regex search with optional fuzzy matching:

```go
rl.Searcher = func(needle string, haystack []string) []string {
    fz, _ := util.DoString(l, "return hilbish.opts.fuzzy")
    fuzz, ok := fz.TryBool()
    if !fuzz || !ok {
        return regexSearcher(needle, haystack)  // default regex search
    }
    matches := fuzzy.Find(needle, haystack)  // sahilm/fuzzy
    suggs := make([]string, 0)
    for _, match := range matches {
        suggs = append(suggs, match.Str)
    }
    return suggs
}
```

Controlled by `hilbish.opts.fuzzy`:
- `false` (default) — regex-based history search via Ctrl-R
- `true` — fuzzy matching using `sahilm/fuzzy` library

## Vim Mode

### Enabling Vim Mode

```lua
hilbish.inputMode('vim')  -- or 'emacs' (default)
```

The Go implementation:

```go
func hlinputMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
    switch mode {
    case "emacs":
        unsetVimMode()
        lr.rl.InputMode = readline.Emacs
    case "vim":
        setVimMode("insert")
        lr.rl.InputMode = readline.Vim
    }
}
```

### Vim Mode Callback

When vim mode changes, the `ViModeCallback` fires:

```go
rl.ViModeCallback = func(mode readline.ViMode) {
    modeStr := ""
    switch mode {
    case readline.VimKeys:      modeStr = "normal"
    case readline.VimInsert:    modeStr = "insert"
    case readline.VimDelete:    modeStr = "delete"
    case readline.VimReplaceOnce, readline.VimReplaceMany:
                                 modeStr = "replace"
    }
    setVimMode(modeStr)
}
```

`setVimMode()` updates `hilbish.vimMode` and emits the `hilbish.vimMode` hook:

```go
func setVimMode(mode string) {
    util.SetField(l, hshMod, "vimMode", rt.StringValue(mode))
    hooks.Emit("hilbish.vimMode", mode)
}
```

### Vim Mode Hook

```lua
bait.catch('hilbish.vimMode', function(mode)
    -- mode: "insert", "normal", "delete", or "replace"
    -- update prompt or status line to show current mode
end)
```

### Vim Action Callback

Vim actions (yank, paste) trigger the `ViActionCallback`:

```go
rl.ViActionCallback = func(action readline.ViAction, args []string) {
    actionStr := ""
    switch action {
    case readline.VimActionPaste: actionStr = "paste"
    case readline.VimActionYank:  actionStr = "yank"
    }
    hooks.Emit("hilbish.vimAction", actionStr, args)
}
```

### Vim Registers

The `hilbish.vim` module provides register access via metatable magic:

```lua
hilbish.vim.registers['a'] = 'some text'  -- set register a
local content = hilbish.vim.registers['a']  -- get register a
```

The metatable delegates to Go readline functions:

```lua
setmetatable(hilbish.vim.registers, {
    __newindex = function(_, k, v)
        hilbish.editor.setVimRegister(k, v)
    end,
    __index = function(_, k)
        return hilbish.editor.getVimRegister(k)
    end
})
```

### Vim Mode Property

```lua
setmetatable(hilbish.vim, {
    __index = function(_, k)
        if k == 'mode' then return hilbish.vimMode end
    end
})
```

`hilbish.vim.mode` returns the current vim mode string (or nil if in emacs mode).

## Syntax Highlighting

The `SyntaxHighlighter` callback enables real-time syntax highlighting:

```go
rl.SyntaxHighlighter = func(line []rune) string {
    highlighter := hshMod.Get(rt.StringValue("highlighter"))
    retVal, err := rt.Call1(l.MainThread(), highlighter,
        rt.StringValue(string(line)))
    // ...
    return highlighted
}
```

To set a custom highlighter, override the function:

```lua
function hilbish.highlighter(line)
    return line:gsub('"%w+"', function(c)
        return lunacolors.green(c)
    end)
end
```

The highlighter receives the raw line and must return a string with ANSI escape sequences for coloring. The `lunacolors` library is the standard way to produce these sequences.

## Inline Hints

The `HintText` callback displays dimmed text after the cursor position:

```go
rl.HintText = func(line []rune, pos int) []rune {
    hinter := hshMod.Get(rt.StringValue("hinter"))
    retVal, err := rt.Call1(l.MainThread(), hinter,
        rt.StringValue(string(line)), rt.IntValue(int64(pos)))
    // ...
    return []rune(hintText)
}
```

To set a custom hinter:

```lua
function hilbish.hinter(line, pos)
    return 'suggested text'
end
```

The hinter is called on every keystroke. Return an empty string to show no hint.

## The `hilbish.editor` API

The `nature/editor.lua` module wraps the Go `readline` module:

```lua
local readline = require 'readline'
local editor = readline.new()
local editorMt = {}

hilbish.editor = {}

function editorMt.__index(_, key)
    if contains({'deleteByAmount', 'getVimRegister', 'getLine',
                 'insert', 'readChar', 'setVimRegister'}, key) then
        -- delegates to editor[key](editor, ...)
    end
    return function(...)
        return editor[key](editor, ...)
    end
end

setmetatable(hilbish.editor, editorMt)
```

### Available Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `deleteByAmount(amount)` | `amount: number` | Delete characters by amount |
| `getLine()` | → `string` | Get current input line |
| `getVimRegister(register)` | `register: string` → `string` | Get vim register content |
| `insert(text)` | `text: string` | Insert text at cursor |
| `readChar()` | → `string` | Read a keystroke (e.g., `"Ctrl-L"`) |
| `setVimRegister(register, text)` | `register: string`, `text: string` | Set vim register content |
| `log(text)` | `text: string` | Print message before prompt without interrupting input |

### readline.new()

Creates a new readline instance. The global `hilbish.editor` uses one shared instance, but `hilbish.read()` creates a separate instance for its own input.

## Prompt Handling

### Left Prompt

```go
func (lr *lineReader) SetPrompt(p string) {
    halfPrompt := strings.Split(p, "\n")
    if len(halfPrompt) > 1 {
        lr.rl.Multiline = true
        lr.rl.SetPrompt(strings.Join(halfPrompt[:len(halfPrompt)-1], "\n"))
        lr.rl.MultilinePrompt = halfPrompt[len(halfPrompt)-1:][0]
    } else {
        lr.rl.Multiline = false
        lr.rl.MultilinePrompt = ""
        lr.rl.SetPrompt(p)
    }
    if initialized && !running {
        lr.rl.RefreshPromptInPlace("")
    }
}
```

Hilbish supports **multiline prompts** — if the prompt string contains `\n`, everything before the last line becomes the static prompt and the last line becomes the editable prompt line.

### Right Prompt

```go
func (lr *lineReader) SetRightPrompt(p string) {
    lr.rl.SetRightPrompt(p)
}
```

### Prompt Verbs

The `fmtPrompt()` function in `main.go` expands prompt verbs:

| Verb | Expansion |
|------|-----------|
| `%d` | Current working directory (abbreviated with `~`) |
| `%D` | Basename of current directory |
| `%h` | Hostname |
| `%u` | Username |

```go
func fmtPrompt(prompt string) string {
    args := []string{
        "d", cwd,
        "D", filepath.Base(cwd),
        "h", host,
        "u", username,
    }
    r := strings.NewReplacer(args...)
    return r.Replace(prompt)
}
```

### Continuation Prompt

```lua
hilbish.multiprompt '-->'
```

Sets the prompt shown when input is incomplete (e.g., missing closing quote). The Go side stores this in `multilinePrompt` and uses it in `continuePrompt()`.

## History

### File History

History is stored in `~/.local/share/hilbish/.hilbish-history` (XDG-compliant). The `fileHistory` struct manages this:

```go
lr.fileHist = newFileHistory(defaultHistPath)
rl.SetHistoryCtrlR("History", &luaHistory{})
rl.HistoryAutoWrite = false
```

`HistoryAutoWrite` is `false` — hilbish manages history writing itself rather than letting readline auto-write.

### History Lua API

The `hilbish.history` module (loaded from `lr.Loader()`):

| Function | Signature | Description |
|----------|-----------|-------------|
| `add(cmd)` | `cmd: string` | Add command to history |
| `all()` | → `table` | Get all history entries |
| `clear()` | | Delete all history |
| `get(index)` | `index: number` → `string` | Get entry by index |
| `size()` | → `number` | Number of history entries |

### Ctrl-R History Search

Pressing Ctrl-R opens an interactive history search. The search behavior depends on `hilbish.opts.fuzzy`:

- `false` (default) — regex-based substring search
- `true` — fuzzy matching via `sahilm/fuzzy`

## Multiline Input

When input ends with `\`, hilbish enters multiline mode:

```go
if strings.HasSuffix(input, "\\") {
    for {
        input, err = continuePrompt(strings.TrimSuffix(input, "\\") + "\n", false)
        if !strings.HasSuffix(input, "\\") {
            break
        }
    }
}
```

The `continuePrompt()` function emits the `"multiline"` hook and uses the `multilinePrompt`:

```go
func continuePrompt(prev string, newline bool) (string, error) {
    hooks.Emit("multiline", nil)
    lr.SetPrompt(multilinePrompt)
    cont, err := lr.Read()
    // ...
    return prev + cont, nil
}
```

## Ctrl-C Handling

When the user presses Ctrl-C:

```go
if err == readline.CtrlC {
    fmt.Println("^C")
    hooks.Emit("hilbish.cancel")
}
```

The `hilbish.cancel` hook is emitted, allowing Lua code to react to input cancellation.

## Readline Instance for `hilbish.read()`

The `hilbish.read()` function creates a **separate** readline instance:

```go
func hlread(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
    lualr := &lineReader{
        rl: readline.NewInstance(),
    }
    lualr.SetPrompt(prompt)
    input, err := lualr.Read()
    // ...
}
```

This separate instance has no shared history with the main editor. It does have completion support but no history search.

## Edge Cases and Known Issues

- **No `.inputrc` equivalent**: Hilbish's readline is not GNU Readline. There is no `.inputrc` file or `bind` command for key rebinding. Key bindings are compiled into the `maxlandon/readline` library.
- **No custom key binding API**: Unlike bash's `bind` builtin or zsh's `bindkey`, hilbish does not expose a Lua API for rebinding keys.
- **ShowVimMode is false**: The `rl.ShowVimMode` is set to `false` — the mode indicator is not shown by readline itself. Users must implement mode display via the `hilbish.vimMode` hook and prompt customization.
- **NoSpace is always true**: The `TabCompleter` hard-codes `NoSpace: true` in every `CompletionGroup`. There is no per-completion control over trailing space.
- **TrimSlash is always false**: Directory completions keep their trailing slash (from `matchPath()`), and `TrimSlash` is `false` in the completion group.

## References

- Hilbish readline API: <https://hilbish.sammyette.party/docs/api/readline>
- Hilbish vim mode: <https://hilbish.sammyette.party/docs/vim-mode>
- Source: `rl.go` — Go readline wrapper and all callbacks
- Source: `nature/editor.lua` — Lua editor wrapper
- Source: `nature/vim.lua` — Vim register and mode access
- Go library: `github.com/maxlandon/readline`

## Related Skills

- [Completion](completion.md) — How the TabCompleter callback bridges to Lua completions
- [Language & API](language.md) — The Lua API, lunacolors, and bait event system
- [Startup & Configuration](startup-config.md) — How prompts and options are configured
