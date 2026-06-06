# Hilbish Completion System

In-depth reference for hilbish's completion system — the `hilbish.completions` module, completion groups (grid/list), the `TabCompleter` bridge, and how external tools like carapace can integrate.

## Overview

Hilbish's completion system is a **two-layer architecture**: a Go layer (`complete.go`) that provides low-level binary/file/directory completion and a Lua layer (`nature/completions/init.lua`) that implements the completion handler and per-command completions. Unlike bash's `complete` builtin or zsh's `compadd`, hilbish exposes completion registration entirely through the **Lua API** via `hilbish.completions.add()`.

The completion pipeline:

1. User presses **Tab**
2. Go `TabCompleter` callback fires (in `rl.go`)
3. Go calls `hilbish.completions.handler(line, pos)` in Lua
4. Lua handler resolves aliases, splits the line, and dispatches
5. If a command-specific completer exists, it is called via `hilbish.completions.call()`
6. Otherwise, file completions are returned as fallback
7. Lua returns `{completionGroups}, prefix` to Go
8. Go converts Lua tables to `readline.CompletionGroup` structs
9. Readline renders the completion menu

## Source Files

| File | Language | Purpose |
|------|----------|---------|
| `complete.go` | Go | Low-level completion functions: `binaryComplete`, `fileComplete`, `dirComplete`, `matchPath`, `escapeFilename`, Lua module loader |
| `rl.go` | Go | `TabCompleter` callback: bridges Lua completion handler to readline `CompletionGroup` structs |
| `nature/completions/init.lua` | Lua | Default `hilbish.completions.handler`, auto-loads per-command completions |
| `nature/completions/cd.lua` | Lua | Completion for `cd` command (directories only) |
| `nature/completions/sudo.lua` | Lua | Completion for `sudo` command (flags + command delegation) |

## The `hilbish.completions` Module (Go Layer)

The Go `completionLoader()` function creates the `hilbish.completions` Lua table with these exports:

| Function | Signature | Description |
|----------|-----------|-------------|
| `add` | `add(scope, cb)` | Registers a completion handler for a scope |
| `bins` | `bins(query, ctx, fields) → entries, prefix` | Returns matching binaries from `$PATH` |
| `files` | `files(query, ctx, fields) → entries, prefix` | Returns matching file paths |
| `dirs` | `dirs(query, ctx, fields) → entries, prefix` | Returns matching directory paths |
| `call` | `call(name, query, ctx, fields) → groups, prefix` | Calls a registered completer by name |
| `handler` | `handler(line, pos)` | The main completion handler (overridable) |

### `hilbish.completions.add(scope, cb)`

Registers a completion callback for a scope. The scope follows the pattern `command.<cmd>`:

```lua
hilbish.completions.add('command.git', function(query, ctx, fields)
    -- completion logic
    return {compGroup}, prefix
end)
```

Internally, the Go function `hcmpAdd` stores the callback in the `luaCompletions` map:

```go
var luaCompletions = map[string]*rt.Closure{}

func hcmpAdd(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
    scope, cb, err := util.HandleStrCallback(t, c)
    if err != nil {
        return nil, err
    }
    luaCompletions[scope] = cb
    return c.Next(), nil
}
```

### `hilbish.completions.bins(query, ctx, fields)`

Returns executables in `$PATH` matching the query. The Go implementation (`binaryComplete`):

1. If query starts with `./`, `../`, `/`, or `~/` — searches the filesystem for executables
2. Otherwise — scans each directory in `$PATH` using `filepath.Glob(dir + query + "*")`
3. Checks execute permissions with `util.FindExecutable()`
4. Adds Lua-registered commander commands (`cmds.Commands`)
5. Deduplicates results

Returns a Lua table (1-indexed) of completion strings and the prefix string.

### `hilbish.completions.files(query, ctx, fields)`

Returns file/directory matches. The Go implementation (`fileComplete`):

1. Splits the context line using `splitForFile()` to find the last path component
2. Calls `matchPath()` to glob the directory

### `hilbish.completions.dirs(query, ctx, fields)`

Returns directory-only matches. Filters `fileComplete` results to only include directories:

```go
func dirComplete(query, ctx string, fields []string) ([]string, string) {
    fileCompletions, filePref := matchPath(path)
    for _, f := range fileCompletions {
        fullPath, _ := filepath.Abs(util.ExpandHome(query + strings.TrimPrefix(f, filePref)))
        fi, err := os.Stat(fullPath)
        if err != nil { continue }
        if fi.IsDir() {
            completions = append(completions, f)
        }
    }
    return completions, query
}
```

### `hilbish.completions.call(name, query, ctx, fields)`

Calls a registered completer by its scope name. Used to delegate from one completer to another:

```lua
-- In sudo completion: delegate to the target command's completer
return hilbish.completions.call('command.' .. fields[2], query, ctx, fields)
```

The Go implementation looks up the callback in `luaCompletions` and calls it:

```go
func hcmpCall(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
    completer, err := c.StringArg(0)
    // ...
    if completecb, ok = luaCompletions[completer]; !ok {
        return nil, errors.New("completer " + completer + " does not exist")
    }
    err = rt.Call(l.MainThread(), rt.FunctionValue(completecb),
        []rt.Value{rt.StringValue(query), rt.StringValue(ctx), rt.TableValue(fields)}, cont)
    // ...
}
```

### `hilbish.completions.handler(line, pos)`

The main completion handler. This is the function that the Go `TabCompleter` calls. It can be **overridden** by assigning a new function:

```lua
function hilbish.completions.handler(line, pos)
    -- custom completion logic
end
```

## The Default Completion Handler

The default handler is implemented in `nature/completions/init.lua`:

```lua
function hilbish.completions.handler(line, pos)
    if type(line) ~= 'string' then error '#1 must be a string' end
    if type(pos) ~= 'number' then error '#2 must be a number' end

    -- trim leading whitespace
    local ctx = line:gsub('^%s*(.-)$', '%1')
    if ctx:len() == 0 then return {}, '' end

    -- resolve aliases
    local res = hilbish.aliases.resolve(ctx)
    local resFields = string.split(res, ' ')
    local fields = string.split(ctx, ' ')
    if #fields > 1 and #resFields > 1 then
        fields = resFields
    end
    local query = fields[#fields]

    if #fields == 1 then
        -- First word: complete command names
        local comps, pfx = hilbish.completions.bins(query, ctx, fields)
        local compGroup = { items = comps, type = 'grid' }
        return {compGroup}, pfx
    else
        -- Subsequent words: try command-specific completer, fall back to files
        local ok, compGroups, pfx = pcall(hilbish.completions.call,
            'command.' .. fields[1], query, ctx, fields)
        if ok then
            return compGroups, pfx
        end

        local comps, pfx = hilbish.completions.files(query, ctx, fields)
        local compGroup = { items = comps, type = 'grid' }
        return {compGroup}, pfx
    end
end
```

Key behaviors:

- **Alias resolution**: The handler resolves aliases before dispatching, so `ga` (alias for `git add`) triggers the `command.git` completer
- **First word**: Always completes binary names from `$PATH`
- **Subsequent words**: Tries `command.<first_field>` completer first, falls back to file completion
- **Error handling**: Uses `pcall` so a missing completer doesn't crash the shell

## Completion Callback Signature

Every completion callback receives three parameters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | The text being completed (last field) |
| `ctx` | string | The entire command line |
| `fields` | table | The `ctx` split by spaces (1-indexed Lua table) |

The callback must return:

```lua
return {completionGroup1, completionGroup2, ...}, prefix
```

- **First return**: A table of completion group tables
- **Second return**: A prefix string (typically the `query` value)

## Completion Groups

A completion group is a Lua table with two keys:

- `type` (string): `"grid"` or `"list"`
- `items` (table): The completion items

### Grid Type

The simplest type. Items are displayed in a grid layout based on their width:

```lua
{
    items = {'just', 'a bunch', 'of items', 'here', 'hehe'},
    type = 'grid'
}
```

Grid items are plain strings. No descriptions, aliases, or custom display.

### List Type

Displays items in a list format with optional descriptions, aliases, and styled display text:

```lua
{
    items = {
        ['--flag'] = {
            description = 'this flag does a thing',
            alias = '--alias',
            display = lunacolors.format('--{blue}fl{red}ag')
        },
        ['--flag2'] = {
            'make pizza',       -- description (1st positional)
            '--pizzuh',         -- alias (2nd positional)
            display = lunacolors.yellow '--pizzuh'
        },
        '--flag3'              -- simple string item
    },
    type = 'list'
}
```

List item formats:

| Format | Description |
|--------|-------------|
| `['name'] = 'string'` | Simple string item (same as positional) |
| `['name'] = {description, alias}` | Positional: 1st = description, 2nd = alias |
| `['name'] = {description = '...', alias = '...', display = '...'}` | Named keys |
| `'name'` (as value, not key) | Simple string item without description |

### Mixed Groups

Multiple groups of different types can be returned simultaneously:

```lua
local gridGroup = { items = {'file1', 'file2'}, type = 'grid' }
local listGroup = { items = {['--flag'] = {'description'}}, type = 'list' }
return {gridGroup, listGroup}, prefix
```

## Go TabCompleter Bridge (rl.go)

The `TabCompleter` callback in `rl.go` is the bridge between the Lua completion handler and the Go readline library:

```go
rl.TabCompleter = func(line []rune, pos int, _ readline.DelayedTabContext) (string, []*readline.CompletionGroup) {
    // 1. Call Lua handler
    compHandle := hshMod.Get(rt.StringValue("completions")).AsTable().Get(rt.StringValue("handler"))
    err := rt.Call(l.MainThread(), compHandle, []rt.Value{
        rt.StringValue(string(line)),
        rt.IntValue(int64(pos)),
    }, term)

    // 2. Extract Lua return values
    luaCompGroups := term.Get(0)  // table of completion groups
    luaPrefix := term.Get(1)      // prefix string

    // 3. Convert each Lua group to readline.CompletionGroup
    util.ForEach(groups, func(key rt.Value, val rt.Value) {
        valTbl := val.AsTable()
        luaCompType := valTbl.Get(rt.StringValue("type"))
        luaCompItems := valTbl.Get(rt.StringValue("items"))

        // Build items, descriptions, displays, aliases maps
        // ...

        // Map type string to readline display type
        switch luaCompType.AsString() {
        case "grid":
            dispType = readline.TabDisplayGrid
        case "list":
            dispType = readline.TabDisplayList
        }

        compGroups = append(compGroups, &readline.CompletionGroup{
            DisplayType:  dispType,
            Aliases:      itemAliases,
            Descriptions: itemDescriptions,
            ItemDisplays: itemDisplays,
            Suggestions:  items,
            TrimSlash:    false,
            NoSpace:      true,
        })
    })

    return pfx, compGroups
}
```

### CompletionGroup Fields

The Go `readline.CompletionGroup` struct maps to Lua as follows:

| Go Field | Lua Source | Description |
|----------|-----------|-------------|
| `Suggestions` | `items` keys or values | The completion candidates |
| `Descriptions` | `items[name].description` or `items[name][1]` | Per-item description text |
| `Aliases` | `items[name].alias` or `items[name][2]` | Alternative completion text |
| `ItemDisplays` | `items[name].display` | Custom styled display (lunacolors) |
| `DisplayType` | `type` | `TabDisplayGrid` or `TabDisplayList` |
| `TrimSlash` | Hard-coded `false` | Whether to trim trailing slashes |
| `NoSpace` | Hard-coded `true` | Whether to suppress trailing space after completion |

**Important**: `NoSpace` is always `true` in hilbish — the shell never automatically appends a space after a completion. This is a significant difference from bash/zsh where `nospace` is opt-in.

## File Completion Internals (Go)

### `matchPath(query)`

The core file matching function:

1. Strips leading `"` from query (quoted path)
2. Inverts escape characters via `escapeInvertReplaer`
3. Expands `~` via `util.ExpandHome()`
4. Resolves to absolute path
5. Extracts `baseName` from `filepath.Base(query)`
6. Reads directory entries with `os.ReadDir()`
7. Filters entries by `strings.HasPrefix(file.Name(), baseName)`
8. Appends `/` to directory entries
9. Escapes special characters in results via `escapeFilename()`

### Character Escaping

Hilbish escapes these characters in file completions:

```go
var charEscapeMap = []string{
    "\"", "\\\"",
    "'", "\\'",
    "`", "\\`",
    " ", "\\ ",
    "(", "\\(",
    ")", "\\)",
    "[", "\\[",
    "]", "\\]",
    "$", "\\$",
    "&", "\\&",
    "*", "\\*",
    ">", "\\>",
    "<", "\\<",
    "|", "\\|",
}
```

This is more aggressive than bash's `COMP_WORDBREAKS` approach — hilbish escapes shell metacharacters in the completion output itself rather than relying on the shell to handle word breaks.

### `splitForFile(str)`

Splits a command line for file completion context. Handles:
- Quoted strings (preserves content within `"..."`)
- Escaped spaces (`\ ` are not split points)
- Trailing spaces (produce an empty final element, triggering fresh completion)

## Built-in Command Completions

### cd (nature/completions/cd.lua)

```lua
hilbish.completions.add('command.cd', function(query, ctx, fields)
    local comps, pfx = hilbish.completions.dirs(query, ctx, fields)
    local compGroup = { items = comps, type = 'grid' }
    return {compGroup}, pfx
end)
```

Only completes directories using `hilbish.completions.dirs()`.

### sudo (nature/completions/sudo.lua)

A more complex completer that demonstrates:
- Flag completion with descriptions (list type)
- Command completion (grid type)
- Delegation to other command completers via `hilbish.completions.call()`

```lua
hilbish.completions.add('command.sudo', function(query, ctx, fields)
    table.remove(fields, 1)  -- remove 'sudo' from fields
    local nonflags = table.filter(fields, function(v)
        if v == '' then return false end
        return v:match('^%-') == nil
    end)

    if #fields == 1 or #nonflags == 0 then
        if query:match('^%-') then
            -- Complete sudo flags (list type with descriptions)
            local compFlags = {}
            for flg, flgstuff in pairs(flags) do
                if flg:match('^' .. query) then
                    compFlags[flg] = flgstuff
                end
            end
            local compGroup = { items = compFlags, type = 'list' }
            return {compGroup}, query
        end

        -- Complete commands (grid type)
        local comps, pfx = hilbish.completions.bins(query, ctx, fields)
        local compGroup = { items = comps, type = 'grid' }
        return {compGroup}, pfx
    end

    -- Delegate to the target command's completer
    return hilbish.completions.call('command.' .. fields[2], query, ctx, fields)
end)
```

### Auto-loading

The `nature/completions/init.lua` auto-loads all completion files in its directory:

```lua
local info = debug.getinfo(1)
local commandDir = fs.dir(info.source)
local commands = fs.readdir(commandDir)
for _, command in ipairs(commands) do
    local name = command:gsub('%.lua', '')
    if name ~= 'init' then
        require('nature.completions.' .. name)
    end
end
```

## Overriding the Completion Handler

The entire completion pipeline can be replaced by assigning a new function to `hilbish.completions.handler`:

```lua
function hilbish.completions.handler(line, pos)
    -- Custom logic here
    -- Must return: {completionGroups}, prefix
end
```

**Requirements for a custom handler**:
- Must accept `(line, pos)` — the full command line and cursor position
- Must return a table of completion groups and a prefix string
- Must handle alias resolution (the default handler calls `hilbish.aliases.resolve()`)
- Must handle the first-word (command) vs subsequent-words dispatch

## Carapace Integration

Hilbish is not currently a supported shell in carapace, but integration is possible through the completion handler override mechanism.

### Approach: Override `hilbish.completions.handler`

The most natural integration point is to replace the default handler with one that delegates to carapace:

```lua
function hilbish.completions.handler(line, pos)
    local handle = io.popen('carapace ' .. hilbish.opts.shell .. ' _carapace ' .. string.format('%q', line) .. ' ' .. pos)
    if not handle then
        -- fallback to default
        return {}, ''
    end
    local result = handle:read('*a')
    handle:close()

    -- Parse carapace JSON output and convert to hilbish completion groups
    local data = require('dkjson').decode(result)
    -- ... convert to {groups}, prefix format
end
```

### Approach: Register per-command completers

Alternatively, register carapace as a completer for individual commands:

```lua
local function carapaceComplete(cmd, query, ctx, fields)
    local handle = io.popen(cmd .. ' _carapace hilbish ' .. #fields .. ' ' .. string.format('%q', query))
    -- parse and return completion groups
end

hilbish.completions.add('command.git', function(query, ctx, fields)
    return carapaceComplete('git', query, ctx, fields)
end)
```

### Challenges

1. **No native carapace shell support**: Carapace does not currently have a hilbish formatter. A new `internal/shell/hilbish/` module would need to be added.
2. **JSON parsing**: Hilbish does not include a JSON parser by default. A Lua JSON library (e.g., `dkjson`, `cjson`) would need to be available.
3. **NoSpace is always true**: Hilbish hard-codes `NoSpace: true` in the Go `TabCompleter`. There is no Lua-level control over trailing space behavior.
4. **Style mapping**: Carapace styles would need to be mapped to `lunacolors.format()` syntax for the `display` field in list-type completions.
5. **Description length**: Hilbish has no built-in description truncation. Carapace's `CARAPACE_DESCRIPTION_LENGTH` would need to be handled in the bridge.

## Edge Cases and Known Issues

- **No `map` display type**: The Go code has a commented-out `case "map"` for `readline.TabDisplayMap`, indicating this was planned but not implemented.
- **Positional vs named item keys**: The list type supports both `['--flag'] = {'desc', 'alias'}` (positional) and `['--flag'] = {description='desc', alias='alias'}` (named). The Go code tries positional first, then falls back to named keys.
- **Error handling in TabCompleter**: If the Lua handler returns a non-table for completion groups, the Go code silently returns empty results.
- **Alias resolution in handler**: The default handler resolves aliases, but a custom handler must do this manually.
- **No completion for commander commands**: Lua-registered commands (via `commander.register()`) appear in binary completion but have no automatic argument completion.

## References

- Hilbish completions documentation: <https://hilbish.sammyette.party/docs/features/completions>
- Hilbish `hilbish.completions` API: <https://hilbish.sammyette.party/docs/api/hilbish#complete>
- Source: `complete.go` — Go completion functions and Lua module loader
- Source: `rl.go` — `TabCompleter` callback bridging Lua to readline
- Source: `nature/completions/init.lua` — Default completion handler
- Source: `nature/completions/cd.lua` — cd completion
- Source: `nature/completions/sudo.lua` — sudo completion with delegation

## Related Skills

- [Line Editing](line-editing.md) — How the readline library renders completion menus
- [Language & API](language.md) — The Lua API, bait event system, and runner mode
- [Execution](execution.md) — How commands are executed and job control works
- [Startup & Configuration](startup-config.md) — How hilbish initializes and loads completions
