package main

import (
	"github.com/strangedev/concurrent"
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(10)
	data := "hi there"
	m := concurrent.NewHashMap()
	startTime := time.Now()
	allWritten := make(chan bool)
	numEntries := 100000
	go func(){
		var done chan bool
		defer close(allWritten)
		for i := 0; i < numEntries; i++ {
			done = m.Upsert(i, &data)
		}
		<-done
		allWritten<- true
	}()

	time.Sleep(100)
	allRead := make(chan bool)
	go func(){
		notFound := 0
		defer close(allRead)
		for i := 0; i < numEntries; i++ {
			_, ok := m.Get(i)
			if ! ok {
				notFound++
			}
		}
		fmt.Println(notFound)
		allRead<- true
	}()
	<-allWritten
	endTime := time.Now()
	<-allRead
	fmt.Println(endTime.Sub(startTime).String())
}