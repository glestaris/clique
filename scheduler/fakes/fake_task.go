// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/glestaris/ice-clique/scheduler"
)

type FakeTask struct {
	RunStub        func()
	runMutex       sync.RWMutex
	runArgsForCall []struct{}
	PriorityStub        func() int
	priorityMutex       sync.RWMutex
	priorityArgsForCall []struct{}
	priorityReturns struct {
		result1 int
	}
	StateStub        func() scheduler.TaskState
	stateMutex       sync.RWMutex
	stateArgsForCall []struct{}
	stateReturns struct {
		result1 scheduler.TaskState
	}
}

func (fake *FakeTask) Run() {
	fake.runMutex.Lock()
	fake.runArgsForCall = append(fake.runArgsForCall, struct{}{})
	fake.runMutex.Unlock()
	if fake.RunStub != nil {
		fake.RunStub()
	}
}

func (fake *FakeTask) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *FakeTask) Priority() int {
	fake.priorityMutex.Lock()
	fake.priorityArgsForCall = append(fake.priorityArgsForCall, struct{}{})
	fake.priorityMutex.Unlock()
	if fake.PriorityStub != nil {
		return fake.PriorityStub()
	} else {
		return fake.priorityReturns.result1
	}
}

func (fake *FakeTask) PriorityCallCount() int {
	fake.priorityMutex.RLock()
	defer fake.priorityMutex.RUnlock()
	return len(fake.priorityArgsForCall)
}

func (fake *FakeTask) PriorityReturns(result1 int) {
	fake.PriorityStub = nil
	fake.priorityReturns = struct {
		result1 int
	}{result1}
}

func (fake *FakeTask) State() scheduler.TaskState {
	fake.stateMutex.Lock()
	fake.stateArgsForCall = append(fake.stateArgsForCall, struct{}{})
	fake.stateMutex.Unlock()
	if fake.StateStub != nil {
		return fake.StateStub()
	} else {
		return fake.stateReturns.result1
	}
}

func (fake *FakeTask) StateCallCount() int {
	fake.stateMutex.RLock()
	defer fake.stateMutex.RUnlock()
	return len(fake.stateArgsForCall)
}

func (fake *FakeTask) StateReturns(result1 scheduler.TaskState) {
	fake.StateStub = nil
	fake.stateReturns = struct {
		result1 scheduler.TaskState
	}{result1}
}

var _ scheduler.Task = new(FakeTask)
