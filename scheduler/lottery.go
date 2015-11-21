package scheduler

type LotteryTaskSelector struct {
	Rand RandomIntGenerator
}

func (ts *LotteryTaskSelector) SelectTask(tasks []Task) Task {
	if len(tasks) == 0 {
		return nil
	}

	var prioritiesSum int64 = 0
	for _, task := range tasks {
		prioritiesSum += int64(task.Priority())
	}

	n := ts.Rand.Random(int64(prioritiesSum))
	k := 0
	for _, task := range tasks {
		k += task.Priority()
		if int64(k) > n {
			return task
		}
	}

	return nil
}
