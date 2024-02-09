package main

import (
	"fmt"
	"time"
)

const (
	poolInterval   = 2
	reportInterval = 10
)

func main() {
	//var ms MetricsStorage
	//var poolIntervalDuration = time.Duration(poolInterval) * time.Second
	//var reportIntervalDuration = time.Duration(reportInterval) * time.Second
	quit := make(chan bool)
	quit2 := make(chan bool)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				fmt.Println("hear")
			}
		}
	}()

	go func() {
		for {
			select {
			case <-quit2:
				return
			default:
				fmt.Println("334444")
			}
		}
	}()
	<-quit2
	<-quit

	/*			<-time.After(poolIntervalDuration)
				ms.updateValues()
				fmt.Println(ms)*/

	/*	select {
		case:

		case <-time.After(reportIntervalDuration):
			reportMetric(ms)
		}*/
}

func f() {
	for {
		time.Sleep(time.Duration(2) * time.Second)
		fmt.Println("har")
	}
}
