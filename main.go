package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	serverMux := http.NewServeMux()
	server := http.Server{Handler: serverMux, Addr: ":8080"}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}
