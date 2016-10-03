package experiment_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var cliqueAgentBin string

type sbsState struct {
	CliqueAgentBin string `json:"clique_agent_bin"`
}

func TestExperiment(t *testing.T) {
	RegisterFailHandler(Fail)

	var _ = SynchronizedBeforeSuite(func() []byte {
		var s sbsState

		path, err := gexec.Build(
			"github.com/ice-stuff/clique/cmd/clique-agent",
		)
		Expect(err).NotTo(HaveOccurred())
		s.CliqueAgentBin = path

		c, err := json.Marshal(s)
		Expect(err).NotTo(HaveOccurred())

		return c
	}, func(c []byte) {
		var s sbsState

		Expect(json.Unmarshal(c, &s)).To(Succeed())

		cliqueAgentBin = s.CliqueAgentBin
	})

	var _ = SynchronizedAfterSuite(func() {
	}, func() {
		gexec.CleanupBuildArtifacts()
	})

	RunSpecs(t, "Control script Suite")
}
