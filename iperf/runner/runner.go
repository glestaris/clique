package runner

// #cgo CFLAGS: -I ${SRCDIR}/../vendor/src
// #cgo LDFLAGS: -liperf -L${SRCDIR}/../vendor/src/.libs
// #include "runner.h"
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ice-stuff/clique/transfer"
)

type Measurement struct {
	Bytes   uint64
	Seconds float64
}

type EndReport struct {
	SumReceived Measurement `json:"sum_received"`
	SumSent     Measurement `json:"sum_sent"`
}

type report struct {
	End EndReport
}

func ListenAndServe(cfg ServerConfig) (transfer.TransferResults, error) {
	var res transfer.TransferResults

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return res, fmt.Errorf("creating pipe: %s", err)
	}
	defer pipeR.Close()
	defer pipeW.Close()

	result := C.ir_listen_and_serve(
		cfg.ToIRServerConfig(), C.int(pipeW.Fd()),
	)
	if result != C.IR_RESULT_SUCCESS {
		return res, errors.New(C.GoString(C.ir_strerror(C.IRResult(result))))
	}

	rep := report{}
	if err := json.NewDecoder(pipeR).Decode(&rep); err != nil {
		return res, fmt.Errorf("decoding iperf response: %s", err)
	}

	res.BytesSent = uint32(rep.End.SumReceived.Bytes)
	durStr := fmt.Sprintf("%ss", strconv.FormatFloat(
		rep.End.SumSent.Seconds, 'f', -1, 64,
	))
	res.Duration, err = time.ParseDuration(durStr)
	if err != nil {
		return res, fmt.Errorf("parsing duration: %s", err)
	}

	return res, nil
}

func RunTest(cfg ClientConfig) (transfer.TransferResults, error) {
	var res transfer.TransferResults

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return res, fmt.Errorf("creating pipe: %s", err)
	}
	defer pipeR.Close()
	defer pipeW.Close()

	result := C.ir_run_test(
		cfg.ToIRClientConfig(), C.int(pipeW.Fd()),
	)
	if result != C.IR_RESULT_SUCCESS {
		return res, errors.New(C.GoString(C.ir_strerror(C.IRResult(result))))
	}

	rep := report{}
	if err := json.NewDecoder(pipeR).Decode(&rep); err != nil {
		return res, fmt.Errorf("decoding iperf response: %s", err)
	}

	res.BytesSent = uint32(rep.End.SumSent.Bytes)
	durStr := fmt.Sprintf("%ss", strconv.FormatFloat(
		rep.End.SumSent.Seconds, 'f', -1, 64,
	))
	res.Duration, err = time.ParseDuration(durStr)
	if err != nil {
		return res, fmt.Errorf("parsing duration: %s", err)
	}

	return res, nil
}
