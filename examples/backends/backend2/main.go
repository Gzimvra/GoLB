package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
			// Catch-all for anything else -> 404
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "Hello from Backend 2 on :9002")
	})

	fmt.Println("Backend 2 running on :9002")
	http.ListenAndServe(":9002", nil)
}


