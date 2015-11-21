package runner

import (
	"os"

	"github.com/glestaris/ice-clique/config"
	"github.com/onsi/gomega/gbytes"
)

type ClqProcess struct {
	Process       *os.Process
	Buffer        *gbytes.Buffer
	Config        config.Config
	ConfigDirPath string
}

func NewClqProcess(
	proc *os.Process,
	cfg config.Config,
	cfgDirPath string,
) *ClqProcess {
	return &ClqProcess{
		Process:       proc,
		Config:        cfg,
		ConfigDirPath: cfgDirPath,
	}
}

func (c *ClqProcess) Stop() error {
	if err := c.Process.Kill(); err != nil {
		return err
	}

	if err := os.RemoveAll(c.ConfigDirPath); err != nil {
		return err
	}

	return nil
}
