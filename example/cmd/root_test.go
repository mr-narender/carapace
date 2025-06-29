package cmd

import (
	"os"
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/assert"
	"github.com/carapace-sh/carapace/pkg/sandbox"
	"github.com/carapace-sh/carapace/pkg/style"
)

func testScript(t *testing.T, shell string, file string) {
	if content, err := os.ReadFile(file); err != nil {
		t.Fatal("failed to read fixture file")
	} else {
		rootCmd.InitDefaultHelpCmd()
		s, _ := carapace.Gen(rootCmd).Snippet(shell)
		assert.Equal(t, string(content), s+"\n")
	}
}

func TestBash(t *testing.T) {
	testScript(t, "bash", "./_test/bash.sh")
}

func TestBashBle(t *testing.T) {
	testScript(t, "bash-ble", "./_test/bash-ble.sh")
}

func TestCmdClink(t *testing.T) {
	testScript(t, "cmd-clink", "./_test/cmd-clink.lua")
}

func TestElvish(t *testing.T) {
	testScript(t, "elvish", "./_test/elvish.elv")
}

func TestFish(t *testing.T) {
	testScript(t, "fish", "./_test/fish.fish")
}

func TestNushell(t *testing.T) {
	testScript(t, "nushell", "./_test/nushell.nu")
}

func TestOil(t *testing.T) {
	testScript(t, "oil", "./_test/oil.sh")
}

func TestPowershell(t *testing.T) {
	testScript(t, "powershell", "./_test/powershell.ps1")
}

func TestXonsh(t *testing.T) {
	testScript(t, "xonsh", "./_test/xonsh.py")
}

func TestZsh(t *testing.T) {
	testScript(t, "zsh", "./_test/zsh.sh")
}

func TestRoot(t *testing.T) {
	sandbox.Package(t, "github.com/carapace-sh/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("").
			Expect(carapace.Batch(
				carapace.ActionValuesDescribed(
					"action", "action example",
					"alias", "action example",
					"flag", "flag example",
				).Style(style.Blue).Tag("main commands"),
				carapace.ActionValuesDescribed(
					"modifier", "modifier example",
				).Style(style.Yellow).Tag("modifier commands"),
				carapace.ActionValuesDescribed(
					"plugin", "dynamic plugin command",
				).Style(style.Magenta).Tag("plugin commands"),
				carapace.ActionValuesDescribed(
					"chain", "shorthand chain",
					"compat", "",
					"completion", "Generate the autocompletion script for the specified shell",
					"group", "group example",
					"help", "Help about any command",
					"interspersed", "interspersed example",
					"multiparts", "multiparts example",
					"special", "",
					"subcommand", "subcommand example",
				).Tag("other commands"),
			).ToA())

		s.Run("a").
			Expect(carapace.ActionStyledValuesDescribed(
				"action", "action example", style.Blue,
				"alias", "action example", style.Blue,
			).Tag("main commands"))

		s.Run("action").
			Expect(carapace.ActionStyledValuesDescribed(
				"action", "action example", style.Blue,
			).Tag("main commands"))

		s.Run("-").
			Expect(carapace.Batch(
				carapace.ActionStyledValuesDescribed(
					"-C", "change work directory", style.Blue,
					"-a", "multiflag", style.Blue,
					"-h", "help for example", style.Default,
					"-p", "Help message for persistentFlag", style.Yellow,
					"-t", "Help message for toggle", style.Default,
					"-v", "version for example", style.Default,
				).Tag("shorthand flags"),
				carapace.ActionStyledValuesDescribed(
					"--chdir", "change work directory", style.Blue,
					"--array", "multiflag", style.Blue,
					"--help", "help for example", style.Default,
					"--persistentFlag", "Help message for persistentFlag", style.Yellow,
					"--persistentFlag2", "Help message for persistentFlag2", style.Blue,
					"--toggle", "Help message for toggle", style.Default,
					"--version", "version for example", style.Default,
				).Tag("longhand flags"),
			).ToA().NoSpace('.'))

		s.Run("--").
			Expect(carapace.ActionStyledValuesDescribed(
				"--array", "multiflag", style.Blue,
				"--chdir", "change work directory", style.Blue,
				"--help", "help for example", style.Default,
				"--persistentFlag", "Help message for persistentFlag", style.Yellow,
				"--persistentFlag2", "Help message for persistentFlag2", style.Blue,
				"--toggle", "Help message for toggle", style.Default,
				"--version", "version for example", style.Default,
			).NoSpace('.').Tag("longhand flags"))

		s.Run("--a").
			Expect(carapace.ActionStyledValuesDescribed(
				"--array", "multiflag", style.Blue,
			).NoSpace('.').Tag("longhand flags"))

		s.Run("--array").
			Expect(carapace.ActionStyledValuesDescribed(
				"--array", "multiflag", style.Blue,
			).NoSpace('.').Tag("longhand flags"))

		s.Run("--array", "", "--a").
			Expect(carapace.ActionStyledValuesDescribed(
				"--array", "multiflag", style.Blue,
			).NoSpace('.').Tag("longhand flags"))

		s.Run("-a", "", "--a").
			Expect(carapace.ActionStyledValuesDescribed(
				"--array", "multiflag", style.Blue,
			).NoSpace('.').Tag("longhand flags"))
	})
}

func TestOptarg(t *testing.T) {
	sandbox.Package(t, "github.com/carapace-sh/carapace/example")(func(s *sandbox.Sandbox) {
		s.Run("--persistentFlag=").
			Expect(carapace.ActionValues(
				"p1",
				"p2",
				"p3",
			).Prefix("--persistentFlag=").
				Usage("Help message for persistentFlag"))

		s.Run("--persistentFlag=p").
			Expect(carapace.ActionValues(
				"p1",
				"p2",
				"p3",
			).Prefix("--persistentFlag=").
				Usage("Help message for persistentFlag"))

		s.Run("--toggle=").
			Expect(carapace.ActionStyledValues(
				"true", style.Green,
				"false", style.Red,
			).Prefix("--toggle=").
				Usage("Help message for toggle"))

		s.Run("--toggle=tru").
			Expect(carapace.ActionStyledValues(
				"true", style.Green,
			).Prefix("--toggle=").
				Usage("Help message for toggle"))
	})
}

func TestMergeFlags(t *testing.T) {
	sandbox.Package(t, "github.com/carapace-sh/carapace/example")(func(s *sandbox.Sandbox) {
		s.Env("CARAPACE_MERGEFLAGS", "1")
		s.Run("-").
			Expect(carapace.Batch(
				carapace.ActionStyledValuesDescribed(
					"-C", "change work directory", style.Blue,
					"-a", "multiflag", style.Blue,
					"-h", "help for example", style.Default,
					"-p", "Help message for persistentFlag", style.Yellow,
					"-t", "Help message for toggle", style.Default,
					"-v", "version for example", style.Default,
					"--chdir", "change work directory", style.Blue,
					"--array", "multiflag", style.Blue,
					"--help", "help for example", style.Default,
					"--persistentFlag", "Help message for persistentFlag", style.Yellow,
					"--persistentFlag2", "Help message for persistentFlag2", style.Blue,
					"--toggle", "Help message for toggle", style.Default,
					"--version", "version for example", style.Default,
				).Tag("flags"),
			).ToA().NoSpace('.'))

		s.Env("CARAPACE_MERGEFLAGS", "0")
		s.Run("-").
			Expect(carapace.Batch(
				carapace.ActionStyledValuesDescribed(
					"-C", "change work directory", style.Blue,
					"-a", "multiflag", style.Blue,
					"-h", "help for example", style.Default,
					"-p", "Help message for persistentFlag", style.Yellow,
					"-t", "Help message for toggle", style.Default,
					"-v", "version for example", style.Default,
				).Tag("shorthand flags"),
				carapace.ActionStyledValuesDescribed(
					"--chdir", "change work directory", style.Blue,
					"--array", "multiflag", style.Blue,
					"--help", "help for example", style.Default,
					"--persistentFlag", "Help message for persistentFlag", style.Yellow,
					"--persistentFlag2", "Help message for persistentFlag2", style.Blue,
					"--toggle", "Help message for toggle", style.Default,
					"--version", "version for example", style.Default,
				).Tag("longhand flags"),
			).ToA().NoSpace('.'))
	})
}
