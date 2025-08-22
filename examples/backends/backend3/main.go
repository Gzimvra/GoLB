package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Backend 3 on :9003")
	})

	fmt.Println("Backend 3 running on :9003")
	http.ListenAndServe(":9003", nil)
}
