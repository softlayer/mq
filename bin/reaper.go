package main

import (
	"fmt"
	"time"
)

func main() {
	limiter := time.Tick(10 * time.Millisecond)

	for i := 0; i < 10; i++ {
		<-limiter

		fmt.Println(i)
	}
}
