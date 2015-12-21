package devoops

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/process"
)

type Process struct {
	Pid         int
	ProgramName string
	Args        []string
}

func FindByPid(pid int) (Process, error) {
	ok, err := process.PidExists(int32(pid))
	if err != nil {
		return Process{}, fmt.Errorf("getting process: %s", err)
	}
	if !ok {
		return Process{}, fmt.Errorf("process `%d` does not exist", pid)
	}

	// initialize the process
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return Process{}, wrapProcErr(err)
	}

	cmdline, err := p.Cmdline()
	if err != nil {
		return Process{}, wrapProcErr(err)
	}
	cmdArgs := strings.Split(cmdline, " ")

	return Process{
		Pid:         pid,
		ProgramName: cmdArgs[0],
		Args:        cmdArgs[1:],
	}, nil
}

func wrapProcErr(err error) error {
	return fmt.Errorf("accessing process state: %s", err)
}
