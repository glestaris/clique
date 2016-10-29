package transfer_test

import (
	"math/rand"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/transfer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Roundtrip", func() {
	var (
		logger      *logrus.Logger
		port        uint16
		transferrer *transfer.Transferrer
		server      *transfer.Server
		serverCh    chan bool
	)

	BeforeEach(func() {
		var err error

		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}

		port = uint16(5000 + rand.Intn(101) + GinkgoParallelNode())

		server, err = transfer.NewServer(logger, port)
		Expect(err).NotTo(HaveOccurred())
		serverCh = make(chan bool)

		transferrer = &transfer.Transferrer{Logger: logger}
	})

	JustBeforeEach(func() {
		go func(sever *transfer.Server, c chan bool) {
			defer GinkgoRecover()

			server.Serve()
			close(c)
		}(server, serverCh)
	})

	AfterEach(func() {
		Expect(server.Close()).To(Succeed())
		Eventually(serverCh).Should(BeClosed())
	})

	It("should succeed in transfering to a running server", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}
		_, err := transferrer.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return transfer duration", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}

		for i := 0; i < 5; i++ {
			var startTime, endTime time.Time
			var clientRes transfer.TransferResults
			Expect(runWRetries(
				func() error {
					var err error

					startTime = time.Now()
					clientRes, err = transferrer.Transfer(spec)
					endTime = time.Now()

					return err
				}, transfer.ErrServerIsBusy, 20, 200*time.Millisecond,
			)).To(Succeed())

			duration := endTime.Sub(startTime)
			Expect(clientRes.Duration).To(BeNumerically("<", duration))
		}
	})

	It("should return the correct size", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}

		clientRes, err := transferrer.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(clientRes.BytesSent).To(Equal(spec.Size))
	})

	It("should return the same (non-empty) checksum as the server", func(done Done) {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}

		clientRes, err := transferrer.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())

		serverRes := server.LastTrasfer()

		Expect(clientRes.Checksum).NotTo(BeZero())
		Expect(clientRes.Checksum).To(Equal(serverRes.Checksum))

		close(done)
	}, 5.0)

	Describe("Server#Interrupt", func() {
		It("should return an error", func() {
			server.Interrupt()

			spec := transfer.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: port,
			}

			_, err := transferrer.Transfer(spec)
			Expect(err).To(HaveOccurred())
		})

		Describe("Server#Resume", func() {
			Context("when the server is paused", func() {
				BeforeEach(func() {
					server.Interrupt()
				})

				It("should return ok", func() {
					server.Resume()

					spec := transfer.TransferSpec{
						IP:   net.ParseIP("127.0.0.1"),
						Port: port,
					}

					_, err := transferrer.Transfer(spec)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})

func runWRetries(action func() error, retryOnError error, retries int,
	sleep time.Duration) error {
	for i := 0; i < retries; i++ {
		err := action()
		if err != retryOnError {
			return err
		}

		if sleep != 0 {
			time.Sleep(sleep)
		}
		continue
	}

	return retryOnError
}
