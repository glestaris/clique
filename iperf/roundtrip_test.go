package iperf_test

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/iperf"
	"github.com/ice-stuff/clique/testhelpers"
	"github.com/ice-stuff/clique/transfer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Roundtrip", func() {
	var (
		logger       *logrus.Logger
		receiver     *iperf.Receiver
		sender       *iperf.Sender
		senderConn   net.Conn
		receiverConn net.Conn
	)

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}
		iperfPort := testhelpers.SelectPort(GinkgoParallelNode())

		receiver = iperf.NewReceiver(logger, iperfPort)
		sender = iperf.NewSender(logger)

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
					IP:   net.ParseIP("127.0.0.1"),
					Size: 100 * 1024 * 1024,
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
						IP:   net.ParseIP("127.0.0.1"),
						Size: 100 * 1024 * 1024,
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
			}, 5.0)
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
					IP:   net.ParseIP("127.0.0.1"),
					Size: 100 * 1024 * 1024,
				}

				senderRes, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())

				Eventually(receiverDone).Should(BeClosed())

				// Iperf reported amount of bytes sent/received can differ than
				// requested. A 10% error is used here to verify that despite being
				// different, these numbers are not completely wrong.
				margin := float32(spec.Size) * 0.10
				Expect(senderRes.BytesSent).To(
					BeNumerically("~", spec.Size, margin),
				)
				Expect(senderRes.BytesSent).To(
					BeNumerically("~", receiverRes.BytesSent, margin),
				)
			})

			It("should measure a similar duration with the receiver", func() {
				spec := transfer.TransferSpec{
					IP:   net.ParseIP("127.0.0.1"),
					Size: 100 * 1024 * 1024,
				}

				senderRes, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())

				Eventually(receiverDone).Should(BeClosed())
				Expect(senderRes.Duration).NotTo(BeZero())
				Expect(senderRes.Duration).To(
					BeNumerically("~", receiverRes.Duration, 60000000),
				) // +/- 60000000ns = 60ms
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

			// synchronize to avoid flakes
			Eventually(receiver.IsBusy).Should(BeTrue())
		})

		AfterEach(func() {
			_, err := sender.SendTransfer(transfer.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Size: 100 * 1024 * 1024,
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
						IP:   net.ParseIP("127.0.0.1"),
						Size: 100 * 1024 * 1024,
					}
					_, err := sender.SendTransfer(spec, newSenderConn)
					Expect(err).To(HaveOccurred())
					close(newSenderDone)
				}()

				_, err := receiver.ReceiveTransfer(newReceiverConn)
				Expect(err).To(Equal(iperf.ErrBusy))
				Eventually(newSenderDone).Should(BeClosed())

				close(done)
			}, 5.0)
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
					IP:   net.ParseIP("127.0.0.1"),
					Size: 100 * 1024 * 1024,
				}
				_, err := sender.SendTransfer(spec, newSenderConn)
				Expect(err).To(Equal(iperf.ErrBusy))
				Eventually(newReceiverDone).Should(BeClosed())

				close(done)
			}, 5.0)
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
						IP:   net.ParseIP("127.0.0.1"),
						Size: 100 * 1024 * 1024,
					}
					_, err := sender.SendTransfer(spec, senderConn)
					Expect(err).To(HaveOccurred())
					close(senderDone)
				}()

				_, err := receiver.ReceiveTransfer(receiverConn)
				Expect(err).To(Equal(iperf.ErrBusy))
				Eventually(senderDone).Should(BeClosed())

				close(done)
			}, 5.0)
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
					IP:   net.ParseIP("127.0.0.1"),
					Size: 100 * 1024 * 1024,
				}
				_, err := sender.SendTransfer(spec, senderConn)
				Expect(err).To(Equal(iperf.ErrBusy))
				Eventually(receiverDone).Should(BeClosed())

				close(done)
			}, 5.0)
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
					IP:   net.ParseIP("127.0.0.1"),
					Size: 100 * 1024 * 1024,
				}
				_, err := sender.SendTransfer(spec, senderConn)
				Expect(err).NotTo(HaveOccurred())
				Expect(senderConn.Close()).To(Succeed())
				Eventually(receiverDone).Should(BeClosed())

				close(done)
			}, 5.0)
		})
	})
})
