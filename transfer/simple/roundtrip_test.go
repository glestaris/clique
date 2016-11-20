package simple_test

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/simple"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Roundtrip", func() {
	var (
		logger       *logrus.Logger
		receiver     *simple.Receiver
		sender       *simple.Sender
		senderConn   net.Conn
		receiverConn net.Conn
	)

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}

		receiver = simple.NewReceiver(logger)
		sender = simple.NewSender(logger)

		senderConn, receiverConn = net.Pipe()
	})

	AfterEach(func() {
		Expect(senderConn.Close()).To(Succeed())
		Expect(receiverConn.Close()).To(Succeed())
	})

	Context("when the connection is closed", func() {
		var (
			senderConn   net.Conn
			receiverConn net.Conn
		)

		BeforeEach(func() {
			senderConn, receiverConn = net.Pipe()
			Expect(senderConn.Close()).To(Succeed())
			Expect(receiverConn.Close()).To(Succeed())
		})

		Describe("Receiver.ReceiveTransfer", func() {
			It("returns an error", func() {
				_, err := receiver.ReceiveTransfer(receiverConn)
				Expect(err).To(MatchError(ContainSubstring("on closed pipe")))
			})
		})

		Describe("Sender.SendTransfer", func() {
			It("returns an error", func() {
				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}
				_, err := sender.SendTransfer(spec, senderConn)
				Expect(err).To(MatchError(ContainSubstring("on closed pipe")))
			})
		})
	})

	Context("when no receiver is running", func() {
		Describe("Sender.SendTransfer", func() {
			It("should block until a receiver handles the request", func(done Done) {
				senderConn, receiverConn := net.Pipe()
				senderDone := make(chan struct{})
				go func() {
					defer GinkgoRecover()

					spec := transfer.TransferSpec{
						Size: 10 * 1024,
					}
					_, err := sender.SendTransfer(spec, senderConn)
					Expect(err).NotTo(HaveOccurred())
					Expect(senderConn.Close()).To(Succeed())

					close(senderDone)
				}()

				Consistently(senderDone).ShouldNot(BeClosed())

				_, err := receiver.ReceiveTransfer(receiverConn)
				Expect(err).NotTo(HaveOccurred())
				Eventually(senderDone).Should(BeClosed())

				close(done)
			}, 1.0)
		})
	})

	Context("when a receiver is running", func() {
		var (
			receiverRes  transfer.TransferResults
			receiverDone chan struct{}
		)

		BeforeEach(func() {
			receiverDone = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				var err error
				receiverRes, err = receiver.ReceiveTransfer(receiverConn)
				Expect(err).NotTo(HaveOccurred())
				close(receiverDone)
			}()
		})

		Describe("Sender.SendTransfer", func() {
			It("should send the requested number of bytes", func() {
				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}

				senderRes, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())

				Eventually(receiverDone).Should(BeClosed())
				Expect(senderRes.BytesSent).To(Equal(spec.Size))
				Expect(senderRes.BytesSent).To(Equal(receiverRes.BytesSent))
			})

			It("should have the same checksum with the receiver", func() {
				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}

				senderRes, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())

				Eventually(receiverDone).Should(BeClosed())
				Expect(senderRes.Checksum).NotTo(BeZero())
				Expect(senderRes.Checksum).To(Equal(receiverRes.Checksum))
			})

			It("should measure a similar duration with the receiver", func() {
				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}

				senderRes, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())

				Eventually(receiverDone).Should(BeClosed())
				Expect(senderRes.Duration).NotTo(BeZero())
				Expect(senderRes.Duration).To(
					BeNumerically("~", receiverRes.Duration, 500000),
				) // +/- 50000ns = 0.5ms
			})
		})
	})

	Context("when the receiver is busy", func() {
		var receiverDone chan struct{}

		BeforeEach(func() {
			receiverDone = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				_, err := receiver.ReceiveTransfer(receiverConn)
				Expect(err).NotTo(HaveOccurred())
				close(receiverDone)
			}()
		})

		AfterEach(func() {
			_, err := sender.SendTransfer(transfer.TransferSpec{
				Size: 10 * 1024,
			}, senderConn)
			Expect(err).NotTo(HaveOccurred())
			Expect(senderConn.Close()).To(Succeed())
			Eventually(receiverDone).Should(BeClosed())
		})

		Describe("the second call to Receiver.ReceiveTransfer", func() {
			It("should return ErrBusy", func(done Done) {
				newSenderConn, newReceiverConn := net.Pipe()
				newSenderDone := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					spec := transfer.TransferSpec{
						Size: 10 * 1024,
					}
					_, err := sender.SendTransfer(spec, newSenderConn)
					Expect(err).To(HaveOccurred())
					close(newSenderDone)
				}()

				_, err := receiver.ReceiveTransfer(newReceiverConn)
				Expect(err).To(Equal(simple.ErrBusy))
				Eventually(newSenderDone).Should(BeClosed())

				close(done)
			}, 1.0)
		})

		Describe("the second call to Sender.SendTransfer", func() {
			It("should return ErrBusy", func(done Done) {
				newSenderConn, newReceiverConn := net.Pipe()
				newReceiverDone := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					_, err := receiver.ReceiveTransfer(newReceiverConn)
					Expect(err).To(HaveOccurred())
					close(newReceiverDone)
				}()

				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}
				_, err := sender.SendTransfer(spec, newSenderConn)
				Expect(err).To(Equal(simple.ErrBusy))
				Eventually(newReceiverDone).Should(BeClosed())

				close(done)
			}, 1.0)
		})
	})

	Context("when the receiver is interrupted", func() {
		BeforeEach(func() {
			receiver.Interrupt()
		})

		Describe("Receiver.ReceiveTransfer", func() {
			It("should return ErrBusy", func(done Done) {
				senderConn, receiverConn := net.Pipe()
				senderDone := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					spec := transfer.TransferSpec{
						Size: 10 * 1024,
					}
					_, err := sender.SendTransfer(spec, senderConn)
					Expect(err).To(HaveOccurred())
					close(senderDone)
				}()

				_, err := receiver.ReceiveTransfer(receiverConn)
				Expect(err).To(Equal(simple.ErrBusy))
				Eventually(senderDone).Should(BeClosed())

				close(done)
			}, 1.0)
		})

		Describe("Sender.SendTransfer", func() {
			It("should return ErrBusy", func(done Done) {
				senderConn, receiverConn := net.Pipe()
				receiverDone := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					_, err := receiver.ReceiveTransfer(receiverConn)
					Expect(err).To(HaveOccurred())
					close(receiverDone)
				}()

				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}
				_, err := sender.SendTransfer(spec, senderConn)
				Expect(err).To(Equal(simple.ErrBusy))
				Eventually(receiverDone).Should(BeClosed())

				close(done)
			}, 1.0)
		})

		Context("and then resumed", func() {
			BeforeEach(func() {
				receiver.Resume()
			})

			It("works again", func(done Done) {
				senderConn, receiverConn := net.Pipe()
				receiverDone := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					_, err := receiver.ReceiveTransfer(receiverConn)
					Expect(err).NotTo(HaveOccurred())
					close(receiverDone)
				}()

				spec := transfer.TransferSpec{
					Size: 10 * 1024,
				}
				_, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())
				Eventually(receiverDone).Should(BeClosed())

				close(done)
			}, 1.0)
		})
	})
})
