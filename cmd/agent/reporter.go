package main

import "fmt"

func reportMetric(ms MetricsStorage) {
	fmt.Printf("Poll count: %s", ms.PollCount)
}
