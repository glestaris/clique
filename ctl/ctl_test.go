package experiment_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/glestaris/devoops"
	. "github.com/glestaris/devoops/matchers"
	"github.com/ice-stuff/clique/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/shirou/gopsutil/process"
)

var _ = Describe("Control script", func() {
	var (
		ctlPath string
		wd      string
	)

	BeforeEach(func() {
		var err error

		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		ctlPath = path.Join(cwd, "ctl.sh")
		Expect(ctlPath).To(BeAnExistingFile())

		wd, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(wd)).To(Succeed())

		cleanupCliqueAgent()
	})

	Describe("check", func() {
		var cmd *exec.Cmd

		BeforeEach(func() {
			cmd = exec.Command(
				"bash", "-c", fmt.Sprintf("%s check", ctlPath),
			)
			cmd.Dir = wd
		})

		Context("when there is not pid file", func() {
			It("should fail", func() {
				sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Consistently(sess).ShouldNot(gexec.Exit(0))
			})
		})

		Context("when there is a pid file", func() {
			BeforeEach(func() {
				writePidFile(wd, findUnusedPid())
			})

			Context("and the process exists", func() {
				BeforeEach(func() {
					sess := startCliqueAgent(wd)
					writePidFile(wd, sess.Command.Process.Pid)
				})

				It("should succeed", func() {
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))
				})
			})

			Context("and the process does not exist", func() {
				It("should fail", func() {
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Consistently(sess).ShouldNot(gexec.Exit(0))
				})
			})
		})
	})

	Describe("start", func() {
		var cmd *exec.Cmd

		BeforeEach(func() {
			cmd = exec.Command(
				"bash", "-c", fmt.Sprintf("%s start", ctlPath),
			)
			cmd.Dir = wd
		})

		Context("when the binary exists", func() {
			BeforeEach(func() {
				Expect(exec.Command("cp", cliqueAgentBin, wd).Run()).To(Succeed())
			})

			Context("and the configuration file exists", func() {
				var pid int

				BeforeEach(func() {
					writeConfigFile(wd)

					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))

					pid = readPidFile(wd)
				})

				It("should run the agent", func() {
					proc, err := devoops.FindByPid(pid)
					Expect(err).NotTo(HaveOccurred())

					Expect(proc).To(HaveProgramName("./clique-agent"))
				})

				It("should pass the correct arguments", func() {
					proc, err := devoops.FindByPid(pid)
					Expect(err).NotTo(HaveOccurred())

					Expect(proc).To(HaveArgs("-config=./config.json"))
				})

				It("should pass the LD_LIBRARY_PATH environment variable", func() {
					proc, err := devoops.FindByPid(pid)
					Expect(err).NotTo(HaveOccurred())

					ldLibraryPath, _ := syscall.Getenv("LD_LIBRARY_PATH")
					ldLibraryPath = fmt.Sprintf("%s:%s", wd, ldLibraryPath)
					Expect(proc).To(
						ContainEnv(fmt.Sprintf("LD_LIBRARY_PATH=%s", ldLibraryPath)),
					)
				})

				It("should forward output", func() {
					stdoutPath := path.Join(wd, "stdout.log")
					Expect(stdoutPath).To(BeAnExistingFile())

					stderrPath := path.Join(wd, "stderr.log")
					Expect(stderrPath).To(BeAnExistingFile())
				})

				Context("and trying to recall", func() {
					It("should return an error", func() {
						cmd := exec.Command(
							"bash", "-c", fmt.Sprintf("%s start", ctlPath),
						)
						cmd.Dir = wd

						sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
						Expect(err).NotTo(HaveOccurred())
						Consistently(sess).ShouldNot(gexec.Exit(0))
					})
				})
			})

			Context("and the configuration file does not exist", func() {
				It("should fail", func() {
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Consistently(sess).ShouldNot(gexec.Exit(0))
				})
			})
		})

		Context("and the binary does not exist", func() {
			It("should fail", func() {
				sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Consistently(sess).ShouldNot(gexec.Exit(0))
			})
		})
	})

	Describe("stop", func() {
		var cmd *exec.Cmd

		BeforeEach(func() {
			cmd = exec.Command(
				"bash", "-c", fmt.Sprintf("%s stop", ctlPath),
			)
			cmd.Dir = wd
		})

		Context("when the process is running", func() {
			var pid int

			BeforeEach(func() {
				sess := startCliqueAgent(wd)
				pid = sess.Command.Process.Pid
				writePidFile(wd, pid)
			})

			AfterEach(func() {
				cleanupCliqueAgent()
			})

			It("should stop it", func() {
				sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, "3s").Should(gexec.Exit(0))

				_, err = devoops.FindByPid(pid)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the process is not runnning", func() {
			BeforeEach(func() {
				writePidFile(wd, findUnusedPid())
			})

			It("should fail", func() {
				sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Consistently(sess).ShouldNot(gexec.Exit(0))
			})
		})
	})
})

func writeConfigFile(wd string) {
	data, err := json.Marshal(config.Config{TransferPort: 5000})
	Expect(err).NotTo(HaveOccurred())

	Expect(
		ioutil.WriteFile(path.Join(wd, "config.json"), data, 0700),
	).To(Succeed())
}

func readPidFile(wd string) int {
	f, err := os.Open(path.Join(wd, "clique-agent.pid"))
	Expect(err).NotTo(HaveOccurred())

	pidBytes, err := ioutil.ReadAll(f)
	Expect(err).NotTo(HaveOccurred())

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	Expect(err).NotTo(HaveOccurred())

	return pid
}

func writePidFile(wd string, pid int) {
	pidFilePath := path.Join(wd, "clique-agent.pid")
	Expect(ioutil.WriteFile(
		pidFilePath, []byte(fmt.Sprintf("%d", pid)), 0700,
	)).To(Succeed())
}

func cleanupCliqueAgent() {
	sess, err := gexec.Start(
		exec.Command("pkill", "-9", "clique-agent"),
		GinkgoWriter, GinkgoWriter,
	)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(gexec.Exit())
}

func startCliqueAgent(wd string) *gexec.Session {
	writeConfigFile(wd)

	cmd := exec.Command(
		cliqueAgentBin, fmt.Sprintf("-config=%s/config.json", wd),
	)

	sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return sess
}

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
