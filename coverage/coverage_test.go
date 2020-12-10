package coverage

import (
	"testing"
)

func TestProcessCoverage(t *testing.T) {
	err := ProcessCoverage("pingcap", "tidb")
	if err != nil {
		t.Fatal(err)
	}
}
