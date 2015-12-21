# Devoops: Gomega matchers for ops testing

This is a [Gomega](https://onsi.github.io/gomega/) matchers library that helps
with asserting on the running processes.

This is work in progress and it currently only supports:
  * `devoops.FindByPid` finding a process given its PID. The call will return
    an error if the process with provided PID does not exist.
  * `HaveProgramName` verifies the program name of a `devooops.Process` object.
  * `HaveArgs` asserts on the argument list provided to the process.

## Installation

On most occasions you can install devoops by running:

```bash
go get github.com/glestaris/devoops
```

## Use

A Ginkgo/Gomega test that's using Devoops:

```go
package awesome_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/glestaris/devoops"
	. "github.com/glestaris/devoops/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Control script", func() {
	var wd string

  // Run the control script start command `ctl start` and write the working
  //  directory of the control script in the wd variable. The test depends
  //  on the existance of `agent.pid` file, that contains the pid of the
  //  agent that the control script runs.

	Describe("start", func() {
		var pid int

		BeforeEach(func() {
			f, err := os.Open(path.Join(wd, "agent.pid"))
			Expect(err).NotTo(HaveOccurred())

			pidBytes, err := ioutil.ReadAll(f)
			Expect(err).NotTo(HaveOccurred())

			pid, err = strconv.Atoi(strings.TrimSpace(string(pidBytes)))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should run the awesome agent", func() {
			proc, err := devoops.FindByPid(pid)
			Expect(err).NotTo(HaveOccurred())

			Expect(proc).To(HaveProgramName("awesome-agent"))
		})

		It("should pass the correct arguments", func() {
			proc, err := devoops.FindByPid(pid)
			Expect(err).NotTo(HaveOccurred())

			Expect(proc).To(HaveArgs("-config=config.json"))
		})
	})
})
```
