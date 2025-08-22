package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Backend 2 on :9002")
	})

	fmt.Println("Backend 2 running on :9002")
	http.ListenAndServe(":9002", nil)
}


