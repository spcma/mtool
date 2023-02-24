package main

import "fmt"

func main() {
	aa := AA{}
	//a.A = new(A)
	a := A{}

	aa.A = a
	a.Name = "hello"
	println(a.GetName())
	println(aa.GetName())

	aa.Name = "go"
	println(a.GetName())
	println(aa.GetName())

	fmt.Println(aa)
}

type A struct {
	Name string
}

func (a *A) GetName() string {
	return a.Name
}

type AA struct {
	A
}
