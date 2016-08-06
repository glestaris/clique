package scheduler_test

import (
	"math"

	"github.com/glestaris/clique/scheduler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// This is the critical values of the chi-squared distribution
// with 999 degrees of freedom
const CriticalValue5th = 926.631
const CriticalValue10th = 942.161
const CriticalValue25th = 968.499
const CriticalValue975th = 1088.49

var _ = Describe("Random", func() {
	var (
		max int
		gen scheduler.RandomIntGenerator
	)

	BeforeEach(func() {
		max = 1000
	})

	JustBeforeEach(func() {
		gen = scheduler.NewRandUIG()
	})

	It("should return numbers smaller than max", func() {
		samples := 1000

		for i := 0; i < samples; i++ {
			n := gen.Random(int64(max))
			Expect(n).To(BeNumerically("<", max))
		}
	})

	It("should return uniformly random numbers", func() {
		frequency := make([]int, max)
		samples := 1000000

		for i := 0; i < samples; i++ {
			n := gen.Random(int64(max))
			frequency[n]++

			if i%1000 == 0 {
				gen = scheduler.NewRandUIG()
			}
		}

		statistic := chiSquaredUniform(samples, max, frequency)
		// 75% confidence that the distribution is uniform
		// Expect(statistic).To(BeNumerically("<", CriticalValue25th))
		// 2.5% confidence that the distribution is uniform
		Expect(statistic).To(BeNumerically("<", CriticalValue975th))
	})
})

func chiSquaredUniform(samples, categories int, obsFreq []int) float64 {
	expFreq := float64(samples) / float64(categories)

	sum := 0.0
	for i := 0; i < categories; i++ {
		sum += math.Pow(float64(obsFreq[i])-expFreq, 2) / expFreq
	}

	return sum
}
