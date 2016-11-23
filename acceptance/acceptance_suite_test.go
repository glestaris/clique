package acceptance_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"testing"

	"github.com/ice-stuff/clique/acceptance/runner"
	"github.com/ice-stuff/clique/config"
)

var cliqueAgentBin string

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeEach(func() {
		if os.Getenv("CLIQUE_AGENT_PATH") != "" {
			cliqueAgentBin = os.Getenv("CLIQUE_AGENT_PATH")
		} else {
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			cliqueAgentBin = filepath.Join(wd, "clique-agent")
		}
		Expect(cliqueAgentBin).To(BeARegularFile())
	})

	RunSpecs(t, "Acceptance Suite")
}

func startClique(cfg config.Config, args ...string) (*runner.ClqProcess, error) {
	configFile, err := ioutil.TempFile("", "clique-agent-config")
	if err != nil {
		return nil, err
	}
	configFilePath := configFile.Name()

	encoder := json.NewEncoder(configFile)
	if err := encoder.Encode(cfg); err != nil {
		configFile.Close()
		os.Remove(configFilePath)
		return nil, err
	}
	configFile.Close()

	finalArgs := []string{"-config", configFilePath}
	finalArgs = append(finalArgs, args...)
	cmd := exec.Command(cliqueAgentBin, finalArgs...)

	buffer := gbytes.NewBuffer()
	cmd.Stdout = io.MultiWriter(buffer, GinkgoWriter)
	cmd.Stderr = io.MultiWriter(buffer, GinkgoWriter)

	if err := cmd.Start(); err != nil {
		os.Remove(configFilePath)
		return nil, err
	}

	Eventually(buffer).Should(gbytes.Say("Clique Agent"))

	return &runner.ClqProcess{
		Cmd:           cmd,
		Buffer:        buffer,
		Config:        cfg,
		ConfigDirPath: configFilePath,
	}, nil
}
