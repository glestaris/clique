package api_test

import (
	"net"
	"time"

	"github.com/ice-stuff/clique"
	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/api/fakes"
	"github.com/ice-stuff/clique/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Roundtrip", func() {
	var (
		port uint16

		fakeRegistry        *fakes.FakeRegistry
		fakeTransferCreator *fakes.FakeTransferCreator
		server              *api.Server

		client *api.Client
	)

	BeforeEach(func() {
		port = testhelpers.SelectPort(GinkgoParallelNode())

		fakeRegistry = new(fakes.FakeRegistry)
		fakeTransferCreator = new(fakes.FakeTransferCreator)
		server = api.NewServer(
			port,
			fakeRegistry,
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
	})

	Context("when the server is started", func() {
		var serverChan chan struct{}

		BeforeEach(func() {
			serverChan = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				// it's going to fail...
				server.Serve()
				close(serverChan)
			}()

			Eventually(client.Ping).Should(Succeed())
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

		Context("and another server tries to use the same port", func() {
			AfterEach(func() {
				Expect(server.Close()).To(Succeed())
				Eventually(serverChan).Should(BeClosed())
			})

			It("should return an error", func() {
				server := api.NewServer(port, nil, nil)
				Expect(server.Serve()).NotTo(Succeed())
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

			Describe("GET /transfers/<State>", func() {
				var specA, specB api.TransferSpec

				BeforeEach(func() {
					specA = api.TransferSpec{
						IP:   net.ParseIP("127.0.0.1"),
						Port: 1212,
						Size: 1024,
					}

					specB = api.TransferSpec{
						IP:   net.ParseIP("127.0.0.18"),
						Port: 2424,
						Size: 2048,
					}

					fakeRegistry.TransfersByStateStub = func(
						state api.TransferState,
					) []api.Transfer {
						if state == api.TransferStatePending {
							return []api.Transfer{
								api.Transfer{State: api.TransferStatePending, Spec: specA},
							}
						} else if state == api.TransferStateCompleted {
							return []api.Transfer{
								api.Transfer{State: api.TransferStateCompleted, Spec: specB},
							}
						}

						return []api.Transfer{}
					}
				})

				It("should return the list of transfers", func() {
					transfers, err := client.TransfersByState(api.TransferStatePending)
					Expect(err).NotTo(HaveOccurred())
					Expect(transfers).To(HaveLen(1))
					Expect(transfers[0].Spec).To(Equal(specA))

					transfers, err = client.TransfersByState(api.TransferStateRunning)
					Expect(err).NotTo(HaveOccurred())
					Expect(transfers).To(HaveLen(0))

					transfers, err = client.TransfersByState(api.TransferStateCompleted)
					Expect(err).NotTo(HaveOccurred())
					Expect(transfers).To(HaveLen(1))
					Expect(transfers[0].Spec).To(Equal(specB))
				})
			})

			Describe("GET /version", func() {
				It("should return the correct version", func() {
					v, err := client.Version()
					Expect(err).NotTo(HaveOccurred())

					Expect(v).To(Equal(clique.CliqueAgentVersion))
				})
			})

			Describe("GET /transfer_results", func() {
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
							RTT:       time.Millisecond * 12,
							Time:      t,
						},
						api.TransferResults{
							IP:        net.ParseIP("12.15.12.18"),
							BytesSent: 15 * 1024,
							Checksum:  566124,
							Duration:  time.Second * 29,
							RTT:       time.Millisecond * 17,
							Time:      t,
						},
					}
					fakeRegistry.TransferResultsReturns(res)
				})

				It("should return the registry results", func() {
					recvRes, err := client.TransferResults()
					Expect(err).NotTo(HaveOccurred())

					Expect(recvRes).To(Equal(res))
				})

				It("should call the registry", func() {
					client.TransferResults()

					Expect(
						fakeRegistry.TransferResultsCallCount(),
					).To(Equal(1))
				})
			})

			Describe("GET /transfer_results/<IP>", func() {
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
							RTT:       time.Millisecond * 19,
							Time:      t,
						},
					}
					fakeRegistry.TransferResultsByIPReturns(res)
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
						fakeRegistry.TransferResultsByIPCallCount(),
					).To(Equal(1))
					Expect(
						fakeRegistry.TransferResultsByIPArgsForCall(0),
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
