package registry_test

import (
	"math/rand"
	"net"
	"time"

	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/api/registry"
	"github.com/ice-stuff/clique/api/registry/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Registry", func() {
	var r *registry.Registry

	BeforeEach(func() {
		r = registry.NewRegistry()
	})

	Describe("Transfers", func() {
		Context("when there are not transfers", func() {
			It("should return an empty list", func() {
				Expect(r.Transfers()).To(HaveLen(0))
			})

			Describe("TransferByState", func() {
				It("should return an empty list", func() {
					Expect(r.TransfersByState(api.TransferStatePending)).To(HaveLen(0))

					Expect(r.TransfersByState(api.TransferStateRunning)).To(HaveLen(0))
				})
			})
		})

		Context("when transfers are registered", func() {
			var (
				transferSpecA, transferSpecB api.TransferSpec
				staterA, staterB             *fakes.FakeTransferStater
			)

			BeforeEach(func() {
				transferSpecA = api.TransferSpec{
					IP:   net.ParseIP("127.0.0.12"),
					Port: 1024,
					Size: 2048,
				}
				staterA = new(fakes.FakeTransferStater)
				staterA.TransferStateReturns(api.TransferStateRunning)
				r.RegisterTransfer(transferSpecA, staterA)

				transferSpecB = api.TransferSpec{
					IP:   net.ParseIP("127.0.0.48"),
					Port: 8080,
					Size: 4096,
				}
				staterB = new(fakes.FakeTransferStater)
				staterB.TransferStateReturns(api.TransferStateCompleted)
				r.RegisterTransfer(transferSpecB, staterB)
			})

			It("should return the transfers", func() {
				Expect(r.Transfers()).To(Equal([]api.Transfer{
					api.Transfer{
						Spec:  transferSpecA,
						State: api.TransferStateRunning,
					},
					api.Transfer{
						Spec:  transferSpecB,
						State: api.TransferStateCompleted,
					},
				}))
			})

			Describe("TransfersByState", func() {
				It("should return the transfers", func() {
					Expect(r.TransfersByState(api.TransferStatePending)).To(HaveLen(0))

					transfers := r.TransfersByState(api.TransferStateRunning)
					Expect(transfers).To(Equal([]api.Transfer{
						api.Transfer{
							Spec:  transferSpecA,
							State: api.TransferStateRunning,
						},
					}))

					transfers = r.TransfersByState(api.TransferStateCompleted)
					Expect(transfers).To(Equal([]api.Transfer{
						api.Transfer{
							Spec:  transferSpecB,
							State: api.TransferStateCompleted,
						},
					}))
				})
			})

			Context("when a stater changes state", func() {
				It("returns a new transfer instance", func() {
					Expect(r.Transfers()).To(Equal([]api.Transfer{
						api.Transfer{
							Spec:  transferSpecA,
							State: api.TransferStateRunning,
						},
						api.Transfer{
							Spec:  transferSpecB,
							State: api.TransferStateCompleted,
						},
					}))

					staterA.TransferStateReturns(api.TransferStateCompleted)

					Expect(r.Transfers()).To(Equal([]api.Transfer{
						api.Transfer{
							Spec:  transferSpecA,
							State: api.TransferStateCompleted,
						},
						api.Transfer{
							Spec:  transferSpecB,
							State: api.TransferStateCompleted,
						},
					}))
				})

				Describe("TransfersByState", func() {
					It("returns a new transfer instance", func() {
						transfers := r.TransfersByState(api.TransferStateRunning)
						Expect(transfers).To(Equal([]api.Transfer{
							api.Transfer{
								Spec:  transferSpecA,
								State: api.TransferStateRunning,
							},
						}))

						staterA.TransferStateReturns(api.TransferStateCompleted)

						transfers = r.TransfersByState(api.TransferStateRunning)
						Expect(transfers).To(HaveLen(0))

						transfers = r.TransfersByState(api.TransferStateCompleted)
						Expect(transfers).To(Equal([]api.Transfer{
							api.Transfer{
								Spec:  transferSpecA,
								State: api.TransferStateCompleted,
							},
							api.Transfer{
								Spec:  transferSpecB,
								State: api.TransferStateCompleted,
							},
						}))
					})
				})
			})
		})
	})

	Describe("TransferResults", func() {
		It("should return an empty list", func() {
			Expect(r.TransferResults()).To(BeEmpty())
		})

		Context("when results have been registered", func() {
			var transferResultsList []api.TransferResults

			BeforeEach(func() {
				ips := []net.IP{
					net.ParseIP("129.168.1.20"),
					net.ParseIP("129.168.1.14"),
					net.ParseIP("129.168.1.12"),
					net.ParseIP("129.168.1.17"),
				}

				transferResultsList = []api.TransferResults{}
				for _, ip := range ips {
					res := makeTranaferResults(ip, 20*1024*1024)
					transferResultsList = append(transferResultsList, res)

					r.RegisterResults(ip, res)
				}
			})

			It("should return them", func() {
				res := r.TransferResults()

				Expect(res).To(Equal(transferResultsList))
				// is it in its own array?
				Expect(cap(res)).To(Equal(4))
			})
		})
	})

	Describe("TransferResultsByIP", func() {
		It("should return an empty list", func() {
			Expect(r.TransferResultsByIP(net.ParseIP("127.0.0.1"))).To(BeEmpty())
		})

		Context("when results have been registered", func() {
			var (
				transferResultsList []api.TransferResults
				targetIP            net.IP
			)

			BeforeEach(func() {
				targetIP = net.ParseIP("129.168.1.14")
				ips := []net.IP{
					net.ParseIP("129.168.1.20"),
					targetIP,
					net.ParseIP("129.168.1.12"),
					net.ParseIP("129.168.1.17"),
					targetIP,
				}

				transferResultsList = []api.TransferResults{}
				for _, ip := range ips {
					res := makeTranaferResults(ip, 20*1024*1024)
					if ip.Equal(targetIP) {
						transferResultsList = append(transferResultsList, res)
					}

					r.RegisterResults(ip, res)
				}
			})

			It("should return only the results that match the IP", func() {
				res := r.TransferResultsByIP(targetIP)

				Expect(res).To(Equal(transferResultsList))
				// is it in its own array?
				Expect(cap(res)).To(Equal(2))
			})
		})
	})
})

func makeTranaferResults(ip net.IP, bytesSent uint32) api.TransferResults {
	return api.TransferResults{
		IP:        ip,
		BytesSent: bytesSent,
		Checksum:  uint32(rand.Int31()),
		Duration:  time.Duration(rand.Int63n(1000)) * time.Millisecond,
		Time:      time.Now(),
	}
}
