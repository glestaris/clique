package config_test

import (
	"io/ioutil"
	"os"

	"github.com/glestaris/ice-clique/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		configContents string
		configPath     string
	)

	BeforeEach(func() {
		configContents = ""
	})

	JustBeforeEach(func() {
		f, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		_, err = f.Write([]byte(configContents))
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())

		configPath = f.Name()
	})

	AfterEach(func() {
		Expect(os.Remove(configPath)).To(Succeed())
	})

	Context("when the configuration file does not exist", func() {
		It("should fail", func() {
			_, err := config.NewConfig("/path/to/banana/cfg")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the configuration file is malformed", func() {
		BeforeEach(func() {
			configContents = "{'"
		})

		It("should fail", func() {
			_, err := config.NewConfig(configPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when options are missing", func() {
		BeforeEach(func() {
			configContents = "{}"
		})

		It("should fail", func() {
			_, err := config.NewConfig(configPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the configuration file is valid", func() {
		BeforeEach(func() {
			configContents = `{"transfer_port":5000}`
		})

		It("should succeed", func() {
			cfg, err := config.NewConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.TransferPort).To(BeNumerically("==", 5000))
		})
	})
})
