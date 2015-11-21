package scheduler

import (
	"errors"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pivotal-golang/clock"
)

type TaskState int

const (
	TaskStateReady TaskState = iota
	TaskStateDone
)

func (t TaskState) String() string {
	if t == TaskStateReady {
		return "ready"
	} else if t == TaskStateDone {
		return "done"
	}

	return "unknown"
}

//go:generate counterfeiter . Task
type Task interface {
	Run()
	Priority() int
	State() TaskState
}

type Scheduler interface {
	Schedule(task Task)
	Run()
	Stop() error
}

//go:generate counterfeiter . TaskSelector
type TaskSelector interface {
	SelectTask([]Task) Task
}

type schedulerState int

const (
	schedulerStateIdle schedulerState = iota
	schedulerStateRunning
	schedulerStateStopping
)

type scheduler struct {
	taskSelector TaskSelector

	csSleep time.Duration
	csClock clock.Clock

	tasksList []Task

	state schedulerState

	lock sync.RWMutex

	logger *logrus.Logger
}

func NewScheduler(
	logger *logrus.Logger,
	taskSelector TaskSelector,
	csSleep time.Duration,
	csClock clock.Clock,
) Scheduler {
	return &scheduler{
		taskSelector: taskSelector,

		csSleep: csSleep,
		csClock: csClock,

		tasksList: []Task{},

		state: schedulerStateIdle,

		lock: sync.RWMutex{},

		logger: logger,
	}
}

func (s *scheduler) Schedule(task Task) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.logger.WithFields(logrus.Fields{
		"priority": task.Priority(),
	}).Debug("New task is scheduled")

	s.tasksList = append(s.tasksList, task)
}

func (s *scheduler) Run() {
	s.setState(schedulerStateRunning)

	for {
		if s.tasksLen() > 0 {
			task := s.taskSelector.SelectTask(s.tasks())
			s.logger.WithFields(logrus.Fields{
				"priority": task.Priority(),
			}).Debug("Task is selected to run")

			task.Run()
			s.logger.WithFields(logrus.Fields{
				"priority":  task.Priority(),
				"new_state": task.State(),
			}).Debug("Task finished running for this round")

			if task.State() == TaskStateDone {
				s.removeTask(task)
			}
		}

		if s.isStopping() {
			s.logger.Debug("Scheduling loop is terminating...")
			break
		}

		if s.csSleep > 0 {
			s.csClock.Sleep(s.csSleep)
		}
	}

	s.setState(schedulerStateIdle)
}

func (s *scheduler) setState(schedulerState schedulerState) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.state = schedulerState
}

func (s *scheduler) isRunning() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.state == schedulerStateRunning
}

func (s *scheduler) isStopping() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.state == schedulerStateStopping
}

func (s *scheduler) tasks() []Task {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.tasksList
}

func (s *scheduler) tasksLen() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.tasksList)
}

func (s *scheduler) removeTask(task Task) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var i int
	for i = 0; i < len(s.tasksList); i++ {
		if s.tasksList[i] == task {
			break
		}
	}

	if i == len(s.tasksList) {
		return
	}

	beginning := []Task{}
	if i > 0 {
		beginning = s.tasksList[:i]
	}

	if i+1 < len(s.tasksList) {
		s.tasksList = append(beginning, s.tasksList[i+1:]...)
	} else {
		s.tasksList = beginning
	}
}

func (s *scheduler) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.state == schedulerStateStopping {
		return errors.New("scheduler is already stopping")
	}

	if s.state != schedulerStateRunning {
		return errors.New("scheduler is not running")
	}

	s.state = schedulerStateStopping
	s.logger.Info("Scheduler is stopping...")

	return nil
}
