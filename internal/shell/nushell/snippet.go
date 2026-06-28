package nushell

import (
	"fmt"

	"github.com/carapace-sh/carapace/pkg/uid"
	"github.com/spf13/cobra"
)

// Snippet creates the nushell completion script.
func Snippet(cmd *cobra.Command) string {
	return SnippetSingle(cmd.Name(), false)
}

// SnippetMulti creates a multi-completer nushell completion script.
func SnippetMulti(names []string, defaultName string, snippetFuncs string) string {
	return fmt.Sprintf(`%[4]vlet carapace_%[1]v_completer = {|spans|
    %[2]v $spans.0 _carapace nushell ...$spans | from json
}

mut current = (($env | default {} config).config | default {} completions)
$current.completions = ($current.completions | default {} external)
$current.completions.external = ($current.completions.external
||| default true enable
|||# backwards compatible workaround for default, see nushell #15654
||| upsert completer { if $in == null { $carapace_%[1]v_completer } else { $in } })

$env.config = $current
`, defaultName, uid.Executable(), "", snippetFuncs)
}

// SnippetSingle creates a single-command nushell completion script.
// When explicitCommand is true, the command name is included in the invocation
// (for multi-completer subcommands) and the full config setup is included.
// When false, only the minimal completer function is generated (standalone mode).
func SnippetSingle(command string, explicitCommand bool) string {
	if explicitCommand {
		return fmt.Sprintf(`let carapace_%[2]v_completer = {|spans|
    %[1]v %[2]v _carapace nushell ...$spans | from json
}

mut current = (($env | default {} config).config | default {} completions)
$current.completions = ($current.completions | default {} external)
$current.completions.external = ($current.completions.external
||| default true enable
|||# backwards compatible workaround for default, see nushell #15654
||| upsert completer { if $in == null { $carapace_%[2]v_completer } else { $in } })

$env.config = $current
`, uid.Executable(), command)
	}

	return fmt.Sprintf(`let %v_completer = {|spans| 
    %v _carapace nushell ...$spans | from json
}`, command, uid.Executable())
}
