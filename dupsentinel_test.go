package dupsentinel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	dupsentinel "github.com/gomatic/yze-go-dupsentinel"
)

// TestADuplicatedSentinelIsReportedAndAnAliasIsNot drives the analyzer over a
// fixture holding every shape: a retyped sentinel, one inside a grouped
// declaration, the alias that is the fix, a copy of a message no consumer can
// see, and an ordinary string that merely reads alike.
func TestADuplicatedSentinelIsReportedAndAnAliasIsNot(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dupsentinel.Analyzer, "a")
}

// TestRegistrationIsWellFormed pins the identifiers the suite keys on.
func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, dupsentinel.Registration.Validate())
	assert.Equal(t, "yze/dupsentinel", dupsentinel.Registration.RuleID())
	assert.Same(t, dupsentinel.Analyzer, dupsentinel.Registration.Analyzer)
}
