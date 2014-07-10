package main

import "code.google.com/p/go.net/html"
import "fmt"
import "net/http"

func main() {
	resp, err := http.Get("https://www.airbnb.com/rooms/339470")
	if (err != nil) {
		panic(err)
	}
	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	if (err != nil) {
		panic(err)
	}
	fmt.Println(node.FirstChild)
}
