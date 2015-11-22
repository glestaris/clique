package transfer_test

import (
	"fmt"
	"net"
	"time"

	"github.com/glestaris/ice-clique/transfer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("Serve", func() {
		Context("when the port is too low", func() {
			It("should return an error", func() {
				_, err := transfer.NewServer(16)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Close", func() {
		Context("when close is called twice", func() {
			It("should return an error the second time", func() {
				server, err := transfer.NewServer(8080)
				Expect(err).NotTo(HaveOccurred())
				Expect(server.Close()).To(Succeed())
				Expect(server.Close()).NotTo(Succeed())
			})
		})
	})
})

var _ = Describe("Client", func() {
	Describe("Transfer", func() {
		Context("when there is no server", func() {
			It("should return an error", func() {
				client := transfer.NewClient()
				spec := transfer.TransferSpec{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 12121,
				}
				_, err := client.Transfer(spec)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Roundtrip", func() {
	var (
		port       uint16
		transferer transfer.Transferer
		server     transfer.Server
		serverCh   chan bool
	)

	BeforeEach(func() {
		port = 5000 + uint16(GinkgoParallelNode())

		transferer = transfer.NewClient()
	})

	JustBeforeEach(func() {
		var err error

		server, err = transfer.NewServer(port)
		Expect(err).NotTo(HaveOccurred())
		serverCh = make(chan bool)

		go func(sever transfer.Server, c chan bool) {
			defer GinkgoRecover()

			server.Serve()
			close(c)
		}(server, serverCh)
	})

	AfterEach(func() {
		Expect(server.Close()).To(Succeed())
		Eventually(serverCh).Should(BeClosed())
	})

	It("should listen to the provided port", func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		Expect(err).NotTo(HaveOccurred())
		Expect(conn.Close()).To(Succeed())
	})

	It("should succeed in transfering to a running server", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}
		_, err := transferer.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return transfer duration", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}

		for i := 0; i < 5; i++ {
			startTime := time.Now()
			clientRes, err := transferer.Transfer(spec)
			Expect(err).NotTo(HaveOccurred())
			endTime := time.Now()

			duration := endTime.Sub(startTime)
			Expect(clientRes.Duration).To(BeNumerically("<", duration))

			time.Sleep(time.Millisecond) // wait for server
		}
	})

	It("should return the correct size", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}

		clientRes, err := transferer.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(clientRes.BytesSent).To(Equal(spec.Size))
	})

	It("should return the same (non-empty) checksum as the server", func(done Done) {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024,
		}

		clientRes, err := transferer.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())

		serverRes := server.LastTrasfer()

		Expect(clientRes.Checksum).NotTo(BeZero())
		Expect(clientRes.Checksum).To(Equal(serverRes.Checksum))

		close(done)
	}, 5.0)
})
