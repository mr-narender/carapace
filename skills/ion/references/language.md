# Ion Shell Language Fundamentals

In-depth reference for ion's language syntax — variables, arrays, sigils, type system, methods, quoting, expansion, functions, control flow, and redirection.

## Sigil System

Ion uses **two sigils** to distinguish string and array contexts:

| Sigil | Context | Example |
|-------|---------|---------|
| `$` | String expansion | `echo $name` |
| `@` | Array expansion | `echo @items` |

This is a fundamental departure from POSIX shells where `$` is used for both. The sigil determines how the value is expanded and which methods are available.

## Variables

### Declaration with `let`

```sh
let name = "world"
let count = 42
let pi = 3.14
let flag = true
```

**Key differences from POSIX:**
- Uses `let` keyword (not `VAR=value`)
- Spaces around `=` are required
- No `$` in assignment (only in expansion)

### Multiple Assignment

```sh
let a b = one two
echo $a $b  # one two
```

### Type Annotations

```sh
let a:bool = true
let b:str = "text"
let c:int = 42
let d:float = 3.14
let e:[str] = [one two three]
let f:hmap[str] = [key1=val1 key2=val2]
let g:bmap[str] = [key1=val1 key2=val2]
```

**Supported types:**

| Type | Description | Example |
|------|-------------|---------|
| `str` | String | `let s:str = "hello"` |
| `bool` | Boolean (1/0, true/false) | `let b:bool = true` |
| `int` | Integer | `let n:int = 42` |
| `float` | Floating-point | `let f:float = 3.14` |
| `[T]` | Typed array | `let a:[str] = [a b c]` |
| `hmap[T]` | Hash map | `let m:hmap[str] = [k=v]` |
| `bmap[T]` | B-tree map | `let m:bmap[str] = [k=v]` |

### Variable Mutation

```sh
let count = 0
let count += 1       # increment
let count -= 1       # decrement
let name = "hello"
let name += " world"  # append string
```

### Dropping Variables

```sh
drop name count
```

### Environment Variables

```sh
export EDITOR = vim
export PATH = $PATH:~/bin
```

## Arrays

### Array Declaration

```sh
let items = [ one two three ]
let typed:[int] = [ 1 2 3 ]
```

### Array Expansion

```sh
echo @items          # one two three (each element separate)
echo "@items"        # one two three (expanded in double quotes)
echo '@items'        # @items (literal in single quotes)
```

### Array Indexing

```sh
echo @items[0]       # first element
echo @items[-1]      # last element
```

### Array Concatenation Operators

| Operator | Action | Example |
|----------|--------|---------|
| `++=` | Append | `let arr ++= [4 5]` |
| `::=` | Prepend | `let arr ::= [0 1]` |
| `\\=` | Remove | `let arr \\= [2 3]` |

### Array Copy

```sh
let copy = [ @original ]
```

## Slicing

### String Slicing

```sh
let s = "Hello, World"
echo $s[..5]         # Hello
echo $s[7..]         # World
echo $s[7..12]       # Worl
echo $s[7..=11]      # World (inclusive end)
```

### Array Slicing

```sh
let arr = [ 1 2 3 4 5 ]
echo @arr[0..2]      # 1 2
echo @arr[2..]       # 3 4 5
echo @arr[..=3]      # 1 2 3 4 (inclusive)
```

### Range Syntax

| Syntax | Meaning |
|--------|---------|
| `n..m` | Exclusive end (n to m-1) |
| `n..=m` | Inclusive end (n to m) |
| `n..` | From n to end |
| `..m` | From start to m-1 |
| `..=m` | From start to m |

## String Methods

String methods use the `$` sigil and return strings:

| Method | Description | Example |
|--------|-------------|---------|
| `len` | Grapheme count | `$len("hello")` → `5` |
| `len_bytes` | Byte count | `$len_bytes("héllo")` → `6` |
| `to_lowercase` | Lowercase | `$to_lowercase("HELLO")` → `hello` |
| `to_uppercase` | Uppercase | `$to_uppercase("hello")` → `HELLO` |
| `replace` | Replace all | `$replace(input "old" "new")` |
| `replacen` | Replace first N | `$replacen(input "old" "new" 1)` |
| `regex_replace` | Regex replace | `$regex_replace(input "\\d+" "X")` |
| `reverse` | Reverse graphemes | `$reverse("hello")` → `olleh` |
| `repeat` | Repeat N times | `$repeat("ab" 3)` → `ababab` |
| `find` | Find substring | `$find("hello" "ll")` → `2` |
| `join` | Join array to string | `$join(arr)` or `$join(arr, ", ")` |
| `basename` | Path basename | `$basename("/a/b/c.txt")` → `c.txt` |
| `extension` | File extension | `$extension("/a/b/c.txt")` → `txt` |
| `filename` | Filename without ext | `$filename("/a/b/c.txt")` → `c` |
| `parent` | Parent directory | `$parent("/a/b/c.txt")` → `/a/b` |
| `escape` | Escape string | `$escape(input)` |
| `unescape` | Unescape string | `$unescape(input)` |
| `or` | Fallback if empty | `$or(var "default")` |

**Syntax:**

```sh
echo $len("hello")
echo $replace($string "old" "new")
echo $join(my_array, ", ")
```

Methods can take:
- A variable name (without `$`): `$len(myvar)`
- A literal value: `$len("hello")`
- A nested expression: `$len($(command))`

## Array Methods

Array methods use the `@` sigil and return arrays:

| Method | Description | Example |
|--------|-------------|---------|
| `split` | Split by pattern | `@split("a,b,c" ",")` → `[a b c]` |
| `lines` | Split by newlines | `@lines("a\nb\nc")` → `[a b c]` |
| `chars` | Split into chars | `@chars("abc")` → `[a b c]` |
| `graphemes` | Split into graphemes | `@graphemes("héllo")` |
| `bytes` | Byte values | `@bytes("AB")` → `[65 66]` |
| `reverse` | Reverse array | `@reverse([1 2 3])` → `[3 2 1]` |
| `split_at` | Split at index | `@split_at("hello" 3)` |
| `len` | Element count | `@len([1 2 3])` → `3` |
| `subst` | Default if empty | `@subst(var [default])` |

**Syntax:**

```sh
for elem in @split($string ","); echo $elem; end
let parts = [ @split("a,b,c" ",") ]
```

## Quoting

### Single Quotes

Preserve literal content — no expansion, no escape processing:

```sh
echo 'Hello $world ${variable}'
# Output: Hello $world ${variable}
```

### Double Quotes

Allow variable expansion (`$`, `@`) and escape sequences:

```sh
let name = "world"
echo "Hello $name"
# Output: Hello world

echo "@items"
# Output: one two three (array expanded)
```

### Escape Sequences (with `echo -e`)

| Sequence | Meaning |
|----------|---------|
| `\n` | Newline |
| `\t` | Tab |
| `\\` | Backslash |
| `\a` | Bell |
| `\b` | Backspace |
| `\e` | Escape |
| `\f` | Form feed |
| `\r` | Carriage return |
| `\v` | Vertical tab |
| `\c` | No further output |

### Backslash Escaping

Outside quotes, backslash escapes special characters:

```sh
echo hello\ world     # hello world
echo \$HOME           # $HOME (literal)
```

## Expansions

### Brace Expansion

```sh
echo {1..5}           # 1 2 3 4 5
echo {1..=5}          # 1 2 3 4 5 (inclusive)
echo {5..1}           # 5 4 3 2 1 (descending)
echo {0..3...12}      # 0 3 6 9 12 (step by 3)
echo file.{txt,md}    # file.txt file.md
echo a{b,c}d          # abd acd
```

**Brace expansion does NOT work inside double quotes.**

### Command Substitution

```sh
# String substitution — captures stdout as single string
let date = $(date +%Y)
echo $date

# Array substitution — splits by whitespace
let files = [ @(ls *.txt) ]
echo @files

# Line-split substitution
echo @lines($(ls -1))
```

### Process Substitution

```sh
diff <(sort file1) <(sort file2)
```

### Arithmetic Expansion

```sh
echo $(( 2 + 3 ))           # 5
echo $(( x * x ))           # variable x
let result = $(( 10 / 3 )) # integer division
```

For floating-point math, use the `math` builtin:

```sh
math "3.14 * 2"    # 6.28
math + 3.14 2      # Polish notation
```

### Variable Expansion

```sh
echo $name          # string variable
echo @array         # array variable
echo ${name}        # braced expansion (disambiguate)
echo @{array}       # braced array expansion
```

### Namespace Access

```sh
echo ${env::HOME}       # environment variable
echo ${color::green}    # color namespace
echo ${color::reset}    # reset color
```

## Functions

### Definition

```sh
fn greet name
    echo "Hello, $name!"
end

fn square x:int
    echo $(( x * x ))
end

fn hello name age:int hobbies:[str] -- Greets a person with their info
    echo "$name ($age) has: @hobbies"
end
```

**Features:**
- Type hints on parameters (`x:int`, `hobbies:[str]`)
- Docstrings with `--`
- Single-line syntax: `fn square x:int; echo $(( x * x )); end`

### Calling

```sh
greet John
square 5
hello John 25 [ coding eating ]
```

### Listing Functions

```sh
fn              # list all functions
fn -h           # list with descriptions
```

### Built-in PROMPT Function

```sh
fn PROMPT
    echo -n "${color::green}$USER${color::reset}:$ "
end
```

See [references/startup-config.md](references/startup-config.md) for prompt customization details.

## Control Flow

### If/Else If/Else

```sh
if test $x -eq 5
    echo "x is 5"
else if test $x -gt 5
    echo "x > 5"
else
    echo "x < 5"
end
```

**Key differences from POSIX:**
- Uses `end` instead of `fi`
- Uses `else if` instead of `elif`
- No `then` keyword

### matches (Regex Matching)

```sh
if matches $input '[A-Ma-m]\w+'
    echo "Matches pattern"
end
```

### For Loops

```sh
for element in @array
    echo $element
end

for i in {1..10}
    echo $i
end

# Chunked iteration (multiple variables per step)
for x y z in {1..=10}
    echo $x $y $z
end
```

### While Loops

```sh
let value = 0
while test $value -lt 6
    echo $value
    let value += 1
end
```

### Break and Continue

```sh
for elem in {1..=10}
    if test $elem -eq 5
        break
    end
    if test $((elem % 2)) -eq 1
        continue
    end
end
```

### and / or / not

```sh
test -f file && echo "exists"
test -d dir || echo "not a directory"
not test -f file
```

## Redirection

### Standard Redirection

| Syntax | Meaning |
|--------|---------|
| `cmd > file` | stdout to file (truncate) |
| `cmd >> file` | append stdout to file |
| `cmd < file` | stdin from file |
| `cmd << 'EOF'` | here-document (literal) |
| `cmd <<< "string"` | here-string |

### Stderr-Specific Redirection (Ion Unique)

| Syntax | Meaning |
|--------|---------|
| `cmd ^> file` | stderr to file |
| `cmd ^>> file` | append stderr to file |
| `cmd ^| cmd2` | pipe stderr only |

### Combined Redirection

| Syntax | Meaning |
|--------|---------|
| `cmd &> file` | stdout + stderr to file |
| `cmd &>> file` | append stdout + stderr |
| `cmd &| cmd2` | pipe stdout + stderr |

**The `^>` and `^|` operators are unique to ion** — no other shell provides dedicated stderr piping syntax. This is particularly useful for:

```sh
make ^> errors.log           # capture only stderr
grep pattern ^| less          # page through stderr
make &| tee build.log        # all output through tee
```

## Pipelines

```sh
cmd1 | cmd2              # stdout pipe
cmd1 ^| cmd2             # stderr pipe (ion unique)
cmd1 &| cmd2             # stdout + stderr pipe (ion unique)
cmd1 && cmd2             # cmd2 if cmd1 succeeds
cmd1 || cmd2             # cmd2 if cmd1 fails
cmd1 & cmd2              # cmd1 in background, cmd2 in foreground
cmd1; cmd2               # sequential execution
```

## Globbing

Ion uses the `glob` crate for pathname expansion:

```sh
echo *.txt               # all .txt files
echo **/*.rs             # recursive match
echo [abc]*              # character class
```

Glob options in `filename_completion()`:

| Option | Value |
|--------|-------|
| `case_sensitive` | `true` |
| `require_literal_separator` | `true` |
| `require_literal_leading_dot` | `false` |

## Aliases

```sh
alias ll = "ls -la"
alias gs = "git status"

# Remove alias
unalias ll
```

## Built-in Commands

### File and Variable Commands

| Command | Description |
|---------|-------------|
| `cd [DIR]` | Change directory |
| `echo [-e] [-n] [-s] [STRING]...` | Print text |
| `read VARIABLES...` | Read input into variables |
| `let VAR = VALUE` | Create variable |
| `export VAR = VALUE` | Create environment variable |
| `drop VARIABLES...` | Delete variables |
| `alias NAME = "command"` | Create alias |
| `source FILE` | Execute file in current shell |

### Conditional Commands

| Command | Description |
|---------|-------------|
| `test EXPRESSION` | File and string tests |
| `matches VALUE PATTERN` | Regex match |
| `contains PATTERN VALUES...` | String contains |
| `exists [OPTIONS]` | Check existence |
| `and` | Logical AND (exit status) |
| `or` | Logical OR (exit status) |
| `not` | Negate exit status |
| `is` / `eq` | Equality checks |

### Process Commands

| Command | Description |
|---------|-------------|
| `bg PID` | Background job |
| `fg PID` | Foreground job |
| `jobs` | List jobs |
| `disown [-r\|-h\|-a] [PIDs]` | Remove from shell |
| `wait` | Wait for background jobs |
| `suspend` | Suspend shell (SIGTSTP) |
| `exec [-ch] [cmd]` | Replace shell |
| `eval COMMANDS...` | Evaluate as command |
| `exit [CODE]` | Exit shell |

### Other Commands

| Command | Description |
|---------|-------------|
| `fn [-h]` | List/describe functions |
| `help [BUILTIN]` | Get help |
| `history [options]` | History management |
| `math EXPRESSION` | Floating-point calculator |
| `random [START] [END]` | Random number |
| `which PROGRAM` | Locate command |
| `status [-l\|-i\|-f]` | Shell status |
| `set [-e] [-x] [-o vi\|emacs]` | Set options |
| `dirs` | Print directory stack |
| `pushd DIR` | Push directory |
| `popd` | Pop directory |
| `true` / `false` | Set exit status |

## test Expressions

| Expression | Meaning |
|------------|---------|
| `-e FILE` | File exists |
| `-d FILE` | Is directory |
| `-f FILE` | Is regular file |
| `-r FILE` | Is readable |
| `-w FILE` | Is writable |
| `-x FILE` | Is executable |
| `-s FILE` | Non-empty file |
| `-n STRING` | Non-empty string |
| `-z STRING` | Empty string |
| `STRING1 = STRING2` | String equality |
| `STRING1 != STRING2` | String inequality |
| `INT1 -eq INT2` | Integer equality |
| `INT1 -ne INT2` | Integer inequality |
| `INT1 -lt INT2` | Less than |
| `INT1 -gt INT2` | Greater than |
| `INT1 -le INT2` | Less or equal |
| `INT1 -ge INT2` | Greater or equal |

## References

- [Ion Manual — Variables](https://doc.redox-os.org/ion-manual/variables/00-variables.html)
- [Ion Manual — Arrays](https://doc.redox-os.org/ion-manual/variables/02-arrays.html)
- [Ion Manual — Expansions](https://doc.redox-os.org/ion-manual/expansions/00-expansions.html)
- [Ion Manual — String Methods](https://doc.redox-os.org/ion-manual/expansions/06-stringmethods.html)
- [Ion Manual — Array Methods](https://doc.redox-os.org/ion-manual/expansions/07-arraymethods.html)
- [Ion Manual — Control Flow](https://doc.redox-os.org/ion-manual/control/00-flow.html)
- [Ion Manual — Functions](https://doc.redox-os.org/ion-manual/functions.html)
- [Ion Manual — Builtins](https://doc.redox-os.org/ion-manual/builtins.html)
- [Ion Manual — Slicing](https://github.com/redox-os/ion/blob/master/manual/src/slicing.md)

## Related Skills

- [references/completion.md](references/completion.md) — how ion's completion system works
- [references/execution.md](references/execution.md) — job control, signals, and execution model
- [references/startup-config.md](references/startup-config.md) — initrc, keybindings, and configuration
- [references/line-editing.md](references/line-editing.md) — the Liner library for line editing
