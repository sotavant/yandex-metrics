package utils

const (
	RetriesCount  = 3
	FirstWaitTime = 1
	WaitTimeDiff  = 2
)

func GetRetryWaitTimes() map[int]int {
	interval := make(map[int]int, RetriesCount)
	for i := 1; i <= 3; i++ {
		if i == 1 {
			interval[i] = FirstWaitTime
		} else {
			interval[i] = interval[i-1] + WaitTimeDiff
		}
	}

	return interval
}
