package main

import (
	"fmt"
	"net/http"
)

func handleFunc1(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Hello, World!")
}

func main() {
	http.HandleFunc("/", handleFunc1)
	http.ListenAndServe(":8080", nil)
}
