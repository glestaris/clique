package config_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ice-stuff/clique/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("when the configuration file does not exist", func() {
		It("should fail", func() {
			_, err := config.NewConfig("/path/to/banana/cfg")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the configuraion file exists", func() {
		var (
			cfgPath string
		)

		BeforeEach(func() {
			f, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			cfgPath = f.Name()
		})

		AfterEach(func() {
			Expect(os.Remove(cfgPath)).To(Succeed())
		})

		Context("but it is malformed", func() {
			JustBeforeEach(func() {
				Expect(ioutil.WriteFile(cfgPath, []byte("{'"), 0700)).To(Succeed())
			})

			It("should fail", func() {
				_, err := config.NewConfig(cfgPath)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("and it contains valid json", func() {
			DescribeTable("Validation",
				func(cfg config.Config, valid bool) {
					getConfigFile(cfgPath, cfg)

					_, err := config.NewConfig(cfgPath)
					if valid {
						Expect(err).NotTo(HaveOccurred())
					} else {
						Expect(err).To(HaveOccurred())
					}
				},
				Entry("empty configuation", config.Config{}, false),
				Entry("missing transfer port", config.Config{
					RemoteHosts: []string{"192.168.1.12", "192.168.1.13"},
				}, false),
				Entry("valid configuation", config.Config{
					TransferPort: 5000,
					RemoteHosts:  []string{"192.168.1.12", "192.168.1.13"},
				}, true),
			)

			Describe("Defaults", func() {
				It("should apply the default InitTransferSize", func() {
					getConfigFile(cfgPath, config.Config{
						TransferPort: 5000,
					})

					cfg, err := config.NewConfig(cfgPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(cfg.InitTransferSize).To(BeEquivalentTo(20 * 1024 * 1024))
				})
			})
		})
	})
})

func getConfigFile(cfgPath string, cfg config.Config) {
	contents, err := json.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())

	Expect(ioutil.WriteFile(cfgPath, contents, 0700)).To(Succeed())
}
