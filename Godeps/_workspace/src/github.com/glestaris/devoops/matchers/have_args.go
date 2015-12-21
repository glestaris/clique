package matchers

import (
	"fmt"
	"reflect"

	"github.com/glestaris/devoops"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type haveArgsMatcher struct {
	expected []string
}

func HaveArgs(expected ...string) types.GomegaMatcher {
	return &haveArgsMatcher{
		expected: expected,
	}
}

func (m *haveArgsMatcher) Match(actual interface{}) (success bool, err error) {
	v, ok := actual.(devoops.Process)
	if !ok {
		return false, fmt.Errorf("Can only match against devoops.Process instances")
	}

	return reflect.DeepEqual(v.Args, m.expected), nil
}

func (m *haveArgsMatcher) FailureMessage(actual interface{}) string {
	return format.Message(actual, "to have arguments", m.expected)
}

func (m *haveArgsMatcher) NegatedFailureMessage(actual interface{}) string {
	return format.Message(actual, "not to have arguments", m.expected)
}
