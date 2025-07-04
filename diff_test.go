package carapace

import (
	"testing"

	"github.com/carapace-sh/carapace/pkg/assert"
	"github.com/carapace-sh/carapace/pkg/style"
)

func TestDiff(t *testing.T) {
	original := ActionValues(
		"removed",
		"same",
	)
	new := ActionValues(
		"same",
		"added",
	)

	assert.Equal(t,
		Diff(original, new).Invoke(NewContext()),
		ActionStyledValues(
			"removed", style.Red,
			"same", style.Dim,
			"added", style.Green,
		).Invoke(NewContext()),
	)
}
