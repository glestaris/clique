package transfer_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/transfer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var logger *logrus.Logger

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}
	})

	Describe("Serve", func() {
		var (
			port     uint16
			server   *transfer.Server
			serverCh chan struct{}
		)

		BeforeEach(func() {
			var err error

			port = uint16(5000 + rand.Intn(101) + GinkgoParallelNode())

			server, err = transfer.NewServer(logger, port)
			Expect(err).NotTo(HaveOccurred())

			serverCh = make(chan struct{})
			go func() {
				server.Serve()
				close(serverCh)
			}()
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

		Context("when calling serve in the same port", func() {
			It("should return an error", func() {
				_, err := transfer.NewServer(logger, port)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the server is busy", func() {
			var conn net.Conn

			BeforeEach(func() {
				var err error

				conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
				Expect(err).NotTo(HaveOccurred())
				expectToReadOk(conn)
			})

			AfterEach(func() {
				Expect(conn.Close()).To(Succeed())
			})

			It("should return busy for consecutive transfers", func() {
				busyConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadAll(busyConn)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(data)).To(Equal("i-am-busy"))
			})
		})

		Describe("Interrupt", func() {
			Context("when the server is paused", func() {
				BeforeEach(func() {
					server.Interrupt()
				})

				It("should make server return busy for transfers", func() {
					busyConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
					Expect(err).NotTo(HaveOccurred())

					data, err := ioutil.ReadAll(busyConn)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(data)).To(Equal("i-am-busy"))
				})
			})

			Context("when the server is processing a request", func() {
				var conn net.Conn

				BeforeEach(func() {
					var err error

					conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
					Expect(err).NotTo(HaveOccurred())
					expectToReadOk(conn)
				})

				It("should wait until request is done", func() {
					pauseCh := make(chan bool, 1)
					go func() {
						pauseCh <- true
						server.Interrupt()
						close(pauseCh)
					}()

					Eventually(pauseCh).Should(Receive())
					Expect(pauseCh).NotTo(BeClosed())

					Expect(conn.Close()).To(Succeed())

					Eventually(pauseCh).Should(BeClosed())
				})
			})

			Describe("Resume", func() {
				Context("when a server is paused", func() {
					BeforeEach(func() {
						server.Interrupt()
					})

					It("should make server process requests again", func() {
						server.Resume()

						conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
						Expect(err).NotTo(HaveOccurred())
						expectToReadOk(conn)
					})
				})
			})
		})
	})

	Describe("Close", func() {
		Context("when close is called twice", func() {
			It("should return an error the second time", func() {
				server, err := transfer.NewServer(logger, 8080)
				Expect(err).NotTo(HaveOccurred())
				Expect(server.Close()).To(Succeed())
				Expect(server.Close()).NotTo(Succeed())
			})
		})
	})
})

func expectToReadOk(conn net.Conn) {
	data := make([]byte, 1024)

	n, err := conn.Read(data)
	Expect(err).NotTo(HaveOccurred())

	Expect(string(data[:n])).To(Equal("ok"))
}
