package devoops_test

import (
	"math/rand"
	"os/exec"

	"github.com/glestaris/devoops"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/shirou/gopsutil/process"
)

var _ = Describe("ProcessFinder", func() {
	var sess *gexec.Session

	BeforeEach(func() {
		var err error

		path, err := exec.LookPath("sleep")
		Expect(err).NotTo(HaveOccurred())

		sess, err = gexec.Start(
			&exec.Cmd{
				Path: path,
				Args: []string{"sleep", "100"},
				Env: []string{
					"banana=best_fruit",
					"environ=1",
					"hello=world",
				},
				Dir: "/tmp",
			},
			GinkgoWriter, GinkgoWriter,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		sess.Kill()
		Expect(sess.Wait()).To(gexec.Exit())
	})

	Describe("FindByPid", func() {
		var pid int

		BeforeEach(func() {
			pid = sess.Command.Process.Pid
		})

		It("should return a process with the same pid", func() {
			proc, err := devoops.FindByPid(pid)
			Expect(err).NotTo(HaveOccurred())

			Expect(proc.Pid).To(Equal(pid))
		})

		It("should return a process with the same programm name", func() {
			proc, err := devoops.FindByPid(pid)
			Expect(err).NotTo(HaveOccurred())

			Expect(proc.ProgramName).To(Equal("sleep"))
		})

		It("should return a process with the same arguments", func() {
			proc, err := devoops.FindByPid(pid)
			Expect(err).NotTo(HaveOccurred())

			Expect(proc.Args).To(Equal([]string{"100"}))
		})

		Context("when there is no process identified by this pid", func() {
			It("should return an error", func() {
				pid := findUnusedPid()
				_, err := devoops.FindByPid(pid)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func findUnusedPid() int {
	for {
		pid := rand.Int31()

		ok, err := process.PidExists(pid)
		Expect(err).NotTo(HaveOccurred())
		if !ok {
			return int(pid)
		}
	}
}
