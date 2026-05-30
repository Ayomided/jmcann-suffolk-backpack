package tests

import (
	"testing"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal"
)

func TestHasNameFieldType(t *testing.T) {
	formFieldTypes := internal.FFName | internal.FFAddress | internal.FFEmail
	if formFieldTypes.Has(internal.FFName) != true {
		t.Error("expected form field types to contain Name")
	}
}
