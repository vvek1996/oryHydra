package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Service is healthy from second")
	})

	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
