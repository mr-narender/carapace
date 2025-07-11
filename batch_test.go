package carapace

import (
	"testing"

	"github.com/carapace-sh/carapace/internal/common"
	"github.com/carapace-sh/carapace/pkg/assert"
)

func TestBatch(t *testing.T) {
	b := Batch(
		ActionValues("A", "B"),
		ActionValues("B", "C"),
		ActionValues("C", "D"),
	)
	expected := InvokedAction{
		Action{
			rawValues: common.RawValuesFrom("A", "B", "C", "D"),
		},
	}
	actual := b.Invoke(Context{}).Merge()
	assert.Equal(t, expected, actual)
}

func TestBatchSingle(t *testing.T) {
	b := Batch(
		ActionValues("A", "B"),
	)
	expected := InvokedAction{
		Action{
			rawValues: common.RawValuesFrom("A", "B"),
		},
	}
	actual := b.Invoke(Context{}).Merge()
	assert.Equal(t, expected, actual)
}

func TestBatchNone(t *testing.T) {
	b := Batch()
	expected := InvokedAction{
		Action{
			rawValues: common.RawValuesFrom(),
		},
	}
	actual := b.Invoke(Context{}).Merge()
	assert.Equal(t, expected, actual)
}

func TestBatchToA(t *testing.T) {
	b := Batch(
		ActionValues("A", "B"),
		ActionValues("B", "C"),
		ActionValues("C", "D"),
	)
	expected := InvokedAction{
		Action{
			rawValues: common.RawValuesFrom("A", "B", "C", "D"),
		},
	}
	actual := b.ToA().Invoke(Context{})
	assert.Equal(t, expected, actual)
}
