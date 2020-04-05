package powershell

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/rsteube/carapace/uid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Snippet(cmd *cobra.Command, actions map[string]string) string {
	buf := new(bytes.Buffer)

	var subCommandCases bytes.Buffer
	generatePowerShellSubcommandCases(&subCommandCases, cmd, "")
	fmt.Fprintf(buf, powerShellCompletionTemplate, cmd.Name(), cmd.Name(), subCommandCases.String())

	return buf.String()
}

// TODO only show flag completions when current entry starts with '-', else try powitional completion
// TODO try positional completion if no subcommands
// TODO add empty completion when $completions is empty to prevent fallback to file completion
var powerShellCompletionTemplate = `using namespace System.Management.Automation
using namespace System.Management.Automation.Language
Register-ArgumentCompleter -Native -CommandName '%s' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commandElements = $commandAst.CommandElements
    $state = %v _carapace powershell state $($commandElements| Foreach {$_.Value})
    
    $completions = @(switch ($state) {%s
    })
    $completions.Where{ $_.CompletionText -like "$wordToComplete*" } |
        Sort-Object -Property ListItemText
}`

func generatePowerShellSubcommandCases(out io.Writer, cmd *cobra.Command, previousCommandName string) {
	var cmdName = fmt.Sprintf("%v", uid.Command(cmd))

	fmt.Fprintf(out, "\n        '%s' {", cmdName)

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden {
			usage := escapeStringForPowerShell(flag.Usage)
			if len(flag.Shorthand) > 0 {
				fmt.Fprintf(out, "\n            [CompletionResult]::new('-%s', '-%s', [CompletionResultType]::ParameterName, '%s')", flag.Shorthand, flag.Shorthand, usage)
			}
			fmt.Fprintf(out, "\n            [CompletionResult]::new('--%s', '--%s', [CompletionResultType]::ParameterName, '%s')", flag.Name, flag.Name, usage)
		}
	})

	for _, subCmd := range cmd.Commands() {
		usage := escapeStringForPowerShell(subCmd.Short)
		fmt.Fprintf(out, "\n            [CompletionResult]::new('%s', '%s', [CompletionResultType]::Command, '%s')", subCmd.Name(), subCmd.Name(), usage)
	}

	fmt.Fprint(out, "\n            break\n        }")

	for _, subCmd := range cmd.Commands() {
		generatePowerShellSubcommandCases(out, subCmd, cmdName)
	}
}

func escapeStringForPowerShell(s string) string {
	if s == "" {
		return " " // completion fails if empty (fallback to file completion)
	}
	return strings.Replace(s, "'", "''", -1)
}
