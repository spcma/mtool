package main

import (
	"fmt"
	"time"
)

func main() {
	store := make(chan string)
	go func() {
		for {
			store <- "hello"
			time.Sleep(time.Second)
		}
	}()
	go func() {
		for {
			store <- "hello"
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			msg := <-store
			fmt.Println(msg)
		}
	}()
	go func() {
		for {
			msg := <-store
			fmt.Println(msg)
		}
	}()

	//go func() {
	//	for {
	//		prod(store, "1")
	//		time.Sleep(time.Second)
	//	}
	//}()
	//go func() {
	//	for {
	//		prod(store, "2")
	//		time.Sleep(time.Second)
	//	}
	//}()
	//
	//go func() {
	//	for {
	//		consumer(store, "3")
	//		time.Sleep(time.Second)
	//	}
	//}()
	//
	//go func() {
	//	for {
	//		consumer(store, "4")
	//		time.Sleep(time.Second)
	//	}
	//}()

	stop := make(chan string)
	<-stop
}

func prod(prod chan<- string, no string) {
	fmt.Println(no)
	prod <- "hello"
}

func consumer(cons <-chan string, no string) {
	fmt.Println(no)
	fmt.Println(<-cons)
}
