package scheduler_test

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/clique/scheduler"
	"github.com/glestaris/clique/scheduler/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("Scheduler", func() {
	var (
		sched        *scheduler.Scheduler
		logger       *logrus.Logger
		taskSelector *fakes.FakeTaskSelector
		csSleep      time.Duration
		clk          *fakeclock.FakeClock
	)

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}
		taskSelector = new(fakes.FakeTaskSelector)
		csSleep = time.Millisecond * 5
		t, err := time.Parse(time.RFC3339, "2015-11-24T06:30:00+00:00")
		Expect(err).NotTo(HaveOccurred())
		clk = fakeclock.NewFakeClock(t)

		sched = scheduler.NewScheduler(
			logger,
			taskSelector,
			csSleep,
			clk,
		)
	})

	Describe("Close", func() {
		Context("when the scheduler is not running", func() {
			It("should return an error", func() {
				Expect(sched.Stop()).NotTo(Succeed())
			})
		})

		Context("when the scheduler is running", func() {
			var (
				task    *fakes.FakeTask
				taskCh  chan bool
				schedCh chan struct{}
			)

			BeforeEach(func() {
				taskCh = make(chan bool)

				task = new(fakes.FakeTask)
				task.RunStub = func() {
					taskCh <- true
					<-taskCh
				}

				taskSelector.SelectTaskStub = func(
					tasks []scheduler.Task,
				) scheduler.Task {
					return tasks[0]
				}

				sched.Schedule(task)

				schedCh = make(chan struct{})

				go func(sched *scheduler.Scheduler, c chan struct{}) {
					sched.Run()
					close(c)
				}(sched, schedCh)
			})

			AfterEach(func() {
				close(taskCh)
				Eventually(schedCh).Should(BeClosed())
			})

			Context("and stop is already called (the scheduler is stopping)",
				func() {
					BeforeEach(func() {
						Eventually(taskCh).Should(Receive())

						Expect(sched.Stop()).To(Succeed())
					})

					It("should return an error", func() {
						Expect(sched.Stop()).NotTo(Succeed())
					})
				})
		})
	})

	Describe("Run", func() {
		var schedDone chan struct{}

		BeforeEach(func() {
			schedDone = make(chan struct{})
		})

		JustBeforeEach(func() {
			go func(sched *scheduler.Scheduler, c chan struct{}) {
				defer GinkgoRecover()
				Expect(sched.Run).NotTo(Panic())
				close(c)
			}(sched, schedDone)
			clk.WaitForWatcherAndIncrement(csSleep)
		})

		AfterEach(func() {
			Expect(sched.Stop()).To(Succeed())
			clk.Increment(csSleep)
			Eventually(schedDone).Should(BeClosed())
		})

		Context("when there are no tasks", func() {
			It("should not call the task selector", func() {
				Expect(taskSelector.SelectTaskCallCount()).To(Equal(0))
			})
		})

		Context("when there are tasks", func() {
			var (
				taskA, taskB *fakes.FakeTask
			)

			BeforeEach(func() {
				taskA = new(fakes.FakeTask)
				taskA.PriorityReturns(10)
				taskA.StateReturns(scheduler.TaskStateReady)
				sched.Schedule(taskA)

				taskB = new(fakes.FakeTask)
				taskB.PriorityReturns(5)
				taskB.StateReturns(scheduler.TaskStateReady)
				sched.Schedule(taskB)

				taskSelector.SelectTaskStub = func(
					tasks []scheduler.Task,
				) scheduler.Task {
					return tasks[0]
				}
			})

			It("should call the task selector and run returned task", func() {
				Expect(taskSelector.SelectTaskCallCount()).To(BeNumerically(">", 0))
				Expect(taskSelector.SelectTaskArgsForCall(0)).To(Equal([]scheduler.Task{
					taskA, taskB,
				}))
			})

			It("should run the returned task", func() {
				Expect(taskA.RunCallCount()).To(BeNumerically(">=", 1))
				Expect(taskB.RunCallCount()).To(BeZero())
			})

			It("should honour the context-switch sleep argument", func() {
				clk.WaitForWatcherAndIncrement(csSleep)
				clk.WaitForWatcherAndIncrement(csSleep)
				Expect(taskA.RunCallCount()).To(Equal(3))
			})

			Context("and one is done", func() {
				var (
					taskAState      scheduler.TaskState
					taskAStateMutex sync.Mutex
				)

				BeforeEach(func() {
					taskAState = scheduler.TaskStateReady

					taskA.StateStub = func() scheduler.TaskState {
						taskAStateMutex.Lock()
						defer taskAStateMutex.Unlock()

						return taskAState
					}
				})

				It("should remove it from the list", func() {
					Expect(taskA.RunCallCount()).To(BeNumerically(">=", 1))

					taskAStateMutex.Lock()
					taskAState = scheduler.TaskStateDone
					taskAStateMutex.Unlock()

					clk.WaitForWatcherAndIncrement(csSleep)
					clk.WaitForWatcherAndIncrement(csSleep)

					lastIdx := taskSelector.SelectTaskCallCount()
					Expect(taskSelector.SelectTaskArgsForCall(lastIdx - 1)).To(Equal(
						[]scheduler.Task{taskB},
					))
				})
			})
		})
	})
})
