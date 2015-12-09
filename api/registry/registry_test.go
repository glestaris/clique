package registry_test

import (
	"math/rand"
	"net"
	"time"

	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/api/registry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Registry", func() {
	var r *registry.Registry

	BeforeEach(func() {
		r = registry.NewRegistry()
	})

	Describe("Transfers", func() {
		It("should return an empty list", func() {
			Expect(r.Transfers()).To(BeEmpty())
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

					r.Register(ip, res)
				}
			})

			It("should return them", func() {
				res := r.Transfers()

				Expect(res).To(Equal(transferResultsList))
				// is it in its own array?
				Expect(cap(res)).To(Equal(4))
			})
		})
	})

	Describe("TransfersByIP", func() {
		It("should return an empty list", func() {
			Expect(r.TransfersByIP(net.ParseIP("127.0.0.1"))).To(BeEmpty())
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

					r.Register(ip, res)
				}
			})

			It("should return only the results that match the IP", func() {
				res := r.TransfersByIP(targetIP)

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
