package main

import "fmt"

func main() {
	u := NewUser(UserWithId(1))
	fmt.Println(u.Id)
}

type User struct {
	Id int
}

func NewUser(fs ...UserFunc) *User {
	u := new(User)
	UserFuncs(fs).Apply(u)
	return u
}

type UserFunc func(u *User)
type UserFuncs []UserFunc

func (t UserFuncs) Apply(u *User) {
	for _, userFunc := range t {
		userFunc(u)
	}
}

func UserWithId(id int) UserFunc {
	return func(u *User) {
		u.Id = id
	}
}
