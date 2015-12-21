package matchers

import (
	"fmt"
	"strings"

	"github.com/glestaris/devoops"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type haveProgramNameMatcher struct {
	expected string
}

func HaveProgramName(expected string) types.GomegaMatcher {
	return &haveProgramNameMatcher{
		expected: expected,
	}
}

func (m *haveProgramNameMatcher) Match(actual interface{}) (success bool, err error) {
	v, ok := actual.(devoops.Process)
	if !ok {
		return false, fmt.Errorf("Can only match against devoops.Process instances")
	}

	return strings.Compare(v.ProgramName, m.expected) == 0, nil
}

func (m *haveProgramNameMatcher) FailureMessage(actual interface{}) string {
	return format.Message(actual, "to have program name", m.expected)
}

func (m *haveProgramNameMatcher) NegatedFailureMessage(actual interface{}) string {
	return format.Message(actual, "not to have program name", m.expected)
}
