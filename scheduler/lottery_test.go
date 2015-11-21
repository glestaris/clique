package scheduler_test

import (
	"github.com/glestaris/ice-clique/scheduler"
	"github.com/glestaris/ice-clique/scheduler/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lottery", func() {
	var (
		ts      *scheduler.LotteryTaskSelector
		fakeRIG *fakes.FakeRandomIntGenerator
	)

	BeforeEach(func() {
		fakeRIG = new(fakes.FakeRandomIntGenerator)

		ts = &scheduler.LotteryTaskSelector{
			Rand: fakeRIG,
		}
	})

	It("should allocate fair shares to proceses over time", func() {
		priorities := []int{
			5, 10, 2, 20, 15,
		}
		prioritiesSum := sum(priorities)

		// fake out a fair RNG
		var token int64 = 0
		fakeRIG.RandomStub = func(max int64) int64 {
			token += 1
			if token == max {
				token = 0
			}
			return token
		}

		// make tasks
		fakeTasks := make([]*fakes.FakeTask, 5)
		tasks := make([]scheduler.Task, 5)
		frequency := make([]int, 5)
		for i := 0; i < 5; i++ {
			fakeTasks[i] = new(fakes.FakeTask)
			tasks[i] = fakeTasks[i]

			fakeTasks[i].PriorityReturns(priorities[i])
			fakeTasks[i].StateReturns(scheduler.TaskStateReady)
			idx := i
			fakeTasks[i].RunStub = func() {
				frequency[idx]++
			}
		}

		// run
		for i := 0; i < prioritiesSum*10; i++ {
			t := ts.SelectTask(tasks)
			t.Run()
		}

		// check frequencies
		for i := 0; i < 5; i++ {
			Expect(frequency[i]).To(Equal(priorities[i] * 10))
		}
	})

	Context("when the list of tasks is empty", func() {
		It("should return nil", func() {
			Expect(ts.SelectTask([]scheduler.Task{})).To(BeNil())
		})
	})
})

func sum(arr []int) int {
	ret := 0
	for _, v := range arr {
		ret += v
	}

	return ret
}
