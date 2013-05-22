package main

import "fmt"

func main() {
	r := &Router{}

	r.AddRoute("getQueue", "GET", "/(?P<queue>[a-z]+)")
	r.AddRoute("createQueue", "PUT", "/(?P<queue>[a-z]+)")
	r.AddRoute("deleteQueue", "DELETE", "/(?P<queue>[a-z]+)")

	match := r.Match("DELETE", "/123")

	if match == nil {
		fmt.Println("No match found.")
	}

	fmt.Println(r.Match("GET", "/myqueue"))
}
