package api_test

import (
	"net"
	"time"

	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/api/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Roundtrip", func() {
	var (
		port uint16

		fakeTransferResultsRegistry *fakes.FakeTransferResultsRegistry
		fakeTransferCreator         *fakes.FakeTransferCreator
		server                      *api.Server

		client *api.Client
	)

	BeforeEach(func() {
		port = uint16(6010 + GinkgoParallelNode())

		fakeTransferResultsRegistry = new(fakes.FakeTransferResultsRegistry)
		fakeTransferCreator = new(fakes.FakeTransferCreator)
		server = api.NewServer(
			port,
			fakeTransferResultsRegistry,
			fakeTransferCreator,
		)

		client = api.NewClient("127.0.0.1", port, 0)
	})

	Context("when the server is not running", func() {
		Describe("Client#Ping", func() {
			It("should return an error", func() {
				Expect(client.Ping()).NotTo(Succeed())
			})
		})

		Describe("Server#Serve", func() {
			Context("when the port is invalid", func() {
				It("should return an error", func() {
					server := api.NewServer(12, nil, nil)
					Expect(server.Serve()).NotTo(Succeed())
				})
			})
		})
	})

	Context("when the server is started", func() {
		var (
			serverChan chan bool
		)

		BeforeEach(func() {
			serverChan = make(chan bool)
			go func() {
				defer GinkgoRecover()

				serverChan <- true

				Expect(server.Serve()).To(Succeed())
				close(serverChan)
			}()

			Eventually(serverChan).Should(Receive())
			time.Sleep(time.Millisecond) // hack
			Expect(client.Ping()).To(Succeed())
		})

		Context("and closed immediately", func() {
			BeforeEach(func() {
				Expect(server.Close()).To(Succeed())
				Eventually(serverChan).Should(BeClosed())
			})

			Describe("Server#Close", func() {
				It("should return an error", func() {
					Expect(server.Close()).NotTo(Succeed())
				})
			})

			Describe("Client#Ping", func() {
				It("should return an error", func() {
					Expect(client.Ping()).NotTo(Succeed())
				})
			})
		})

		Context("and it runs", func() {
			AfterEach(func() {
				Expect(server.Close()).To(Succeed())
				Eventually(serverChan).Should(BeClosed())
			})

			Describe("GET /ping", func() {
				It("should succeed", func() {
					Expect(client.Ping()).To(Succeed())
				})
			})

			Describe("GET /transfers", func() {
				var res []api.TransferResults

				BeforeEach(func() {
					t, err := time.Parse(time.RFC3339, "2015-12-20T17:25:12Z")
					Expect(err).NotTo(HaveOccurred())

					res = []api.TransferResults{
						api.TransferResults{
							IP:        net.ParseIP("12.12.12.13"),
							BytesSent: 1024,
							Checksum:  124566,
							Duration:  time.Second * 12,
							Time:      t,
						},
						api.TransferResults{
							IP:        net.ParseIP("12.15.12.18"),
							BytesSent: 15 * 1024,
							Checksum:  566124,
							Duration:  time.Second * 29,
							Time:      t,
						},
					}
					fakeTransferResultsRegistry.TransferResultsReturns(res)
				})

				It("should return the registry results", func() {
					recvRes, err := client.TransferResults()
					Expect(err).NotTo(HaveOccurred())

					Expect(recvRes).To(Equal(res))
				})

				It("should call the registry", func() {
					client.TransferResults()

					Expect(
						fakeTransferResultsRegistry.TransferResultsCallCount(),
					).To(Equal(1))
				})
			})

			Describe("GET /transfers/<IP>", func() {
				var res []api.TransferResults

				BeforeEach(func() {
					t, err := time.Parse(time.RFC3339, "2015-12-20T17:25:12Z")
					Expect(err).NotTo(HaveOccurred())

					res = []api.TransferResults{
						api.TransferResults{
							IP:        net.ParseIP("12.12.12.13"),
							BytesSent: 1024,
							Checksum:  124566,
							Duration:  time.Second * 12,
							Time:      t,
						},
					}
					fakeTransferResultsRegistry.TransferResultsByIPReturns(res)
				})

				It("should return the registry results", func() {
					recvRes, err := client.TransferResultsByIP(net.ParseIP("12.12.12.13"))
					Expect(err).NotTo(HaveOccurred())

					Expect(recvRes).To(Equal(res))
				})

				It("should call the registry with the correct argument", func() {
					ip := net.ParseIP("12.12.12.13")

					client.TransferResultsByIP(ip)

					Expect(
						fakeTransferResultsRegistry.TransferResultsByIPCallCount(),
					).To(Equal(1))
					Expect(
						fakeTransferResultsRegistry.TransferResultsByIPArgsForCall(0),
					).To(Equal(ip))
				})
			})

			Describe("POST /transfers", func() {
				var spec api.TransferSpec

				BeforeEach(func() {
					spec = api.TransferSpec{
						IP:   net.ParseIP("12.13.14.15"),
						Port: 1212,
						Size: 10 * 1024 * 1024,
					}
				})

				It("should succeed", func() {
					Expect(client.CreateTransfer(spec)).To(Succeed())
				})

				It("should call the creator with the correct argument", func() {
					client.CreateTransfer(spec)

					Expect(fakeTransferCreator.CreateCallCount()).To(Equal(1))
					Expect(fakeTransferCreator.CreateArgsForCall(0)).To(Equal(spec))
				})
			})
		})
	})
})
