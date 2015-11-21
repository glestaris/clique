package scheduler

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"os"
	"time"
)

//go:generate counterfeiter . RandomIntGenerator
type RandomIntGenerator interface {
	Random(max int64) int64
}

type randUIG struct {
	rand *rand.Rand
}

func NewRandUIG() RandomIntGenerator {
	return &randUIG{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (u *randUIG) Random(max int64) int64 {
	return u.rand.Int63n(max)
}

type cryptoUIG struct {
	rand io.Reader
}

func NewCryptoUIG() RandomIntGenerator {
	rand, err := os.Open("/dev/urandom")
	if err != nil {
		panic(fmt.Errorf("cannot open `/dev/urandom`: %s", err))
	}

	return &cryptoUIG{
		rand: rand,
	}
}

func (u *cryptoUIG) Random(max int64) int64 {
	biMax := big.NewInt(max)
	n, err := crand.Int(u.rand, biMax)
	if err != nil {
		panic(fmt.Errorf("cannot generate random number: %s", err))
	}

	return n.Int64()
}
