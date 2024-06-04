package utils

const (
	RetriesCount  = 3
	FirstWaitTime = 1
	WaitTimeDiff  = 2
)

// GetRetryWaitTimes получение массива с интервалами повторных запросов
func GetRetryWaitTimes() map[int]int {
	interval := make(map[int]int, RetriesCount)
	const firstIntNum = 1
	for i := 1; i <= 3; i++ {
		if i == firstIntNum {
			interval[i] = FirstWaitTime
		} else {
			interval[i] = interval[i-1] + WaitTimeDiff
		}
	}

	return interval
}
