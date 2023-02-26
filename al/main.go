package main

func main() {
	store := make(chan string, 10)
	store <- ""
}
