// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/glestaris/clique/dispatcher"
	"github.com/glestaris/clique/transfer"
)

type FakeTransferrer struct {
	TransferStub        func(spec transfer.TransferSpec) (transfer.TransferResults, error)
	transferMutex       sync.RWMutex
	transferArgsForCall []struct {
		spec transfer.TransferSpec
	}
	transferReturns struct {
		result1 transfer.TransferResults
		result2 error
	}
}

func (fake *FakeTransferrer) Transfer(spec transfer.TransferSpec) (transfer.TransferResults, error) {
	fake.transferMutex.Lock()
	fake.transferArgsForCall = append(fake.transferArgsForCall, struct {
		spec transfer.TransferSpec
	}{spec})
	fake.transferMutex.Unlock()
	if fake.TransferStub != nil {
		return fake.TransferStub(spec)
	} else {
		return fake.transferReturns.result1, fake.transferReturns.result2
	}
}

func (fake *FakeTransferrer) TransferCallCount() int {
	fake.transferMutex.RLock()
	defer fake.transferMutex.RUnlock()
	return len(fake.transferArgsForCall)
}

func (fake *FakeTransferrer) TransferArgsForCall(i int) transfer.TransferSpec {
	fake.transferMutex.RLock()
	defer fake.transferMutex.RUnlock()
	return fake.transferArgsForCall[i].spec
}

func (fake *FakeTransferrer) TransferReturns(result1 transfer.TransferResults, result2 error) {
	fake.TransferStub = nil
	fake.transferReturns = struct {
		result1 transfer.TransferResults
		result2 error
	}{result1, result2}
}

var _ dispatcher.Transferrer = new(FakeTransferrer)
