package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Backend 1 on :9001")
	})

	fmt.Println("Backend 1 running on :9001")
	http.ListenAndServe(":9001", nil)
}

