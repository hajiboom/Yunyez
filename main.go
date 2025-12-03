package main

import (
	"fmt"
	"sync"
)



func main() {
	ch1 := make(chan bool)
	ch2 := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			<-ch1
			fmt.Printf("%d\n", i)
			ch2<-true
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			<-ch2
			fmt.Printf("%c\n", 'A'+i)
			ch1<-true
		}
	}()

	ch1 <- true
	wg.Wait()
}