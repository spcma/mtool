package main

import (
	"fmt"
	"net/http"
)

// sync.Mutex
// sync.RWMutex

func main() {
	http.HandleFunc("/te", as)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}

}

type A struct{}

func (a A) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("1")
}

func as(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("1")
}
