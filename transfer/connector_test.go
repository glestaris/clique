package transfer_test

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/ice-stuff/clique/transfer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connector", func() {
	var (
		connector  transfer.Connector
		randomPort uint16
	)

	BeforeEach(func() {
		connector = transfer.NewConnector()
		randomPort = uint16(5000 + rand.Intn(101) + GinkgoParallelNode())
	})

	Context("when no server is listening", func() {
		It("should return an error when the listener is not running", func() {
			_, err := connector.Connect(net.ParseIP("127.0.0.1"), randomPort)
			Expect(err).To(MatchError(ContainSubstring("connection refused")))
		})
	})

	Context("when a server is listening", func() {
		var listener net.Listener

		BeforeEach(func() {
			var err error
			listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", randomPort))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(listener.Close()).To(Succeed())
		})

		It("enstablishes a connection", func() {
			go listener.Accept()
			conn, err := connector.Connect(net.ParseIP("127.0.0.1"), randomPort)
			Expect(err).NotTo(HaveOccurred())

			_, err = conn.Write([]byte("hello world"))
			Expect(err).NotTo(HaveOccurred())
			Expect(conn.Close()).To(Succeed())
		})
	})
})
