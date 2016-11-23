package runner

import (
	"os"
	"os/exec"

	"github.com/ice-stuff/clique/config"
	"github.com/onsi/gomega/gbytes"
)

type ClqProcess struct {
	Cmd           *exec.Cmd
	Buffer        *gbytes.Buffer
	Config        config.Config
	ConfigDirPath string
}

func NewClqProcess(
	cmd *exec.Cmd,
	cfg config.Config,
	cfgDirPath string,
) *ClqProcess {
	return &ClqProcess{
		Cmd:           cmd,
		Config:        cfg,
		ConfigDirPath: cfgDirPath,
	}
}

func (c *ClqProcess) Stop() error {
	if err := c.Cmd.Process.Signal(os.Interrupt); err != nil {
		return err
	}

	c.Cmd.Wait()

	if err := os.RemoveAll(c.ConfigDirPath); err != nil {
		return err
	}

	return nil
}
