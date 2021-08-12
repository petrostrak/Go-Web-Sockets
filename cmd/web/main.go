package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	mux := routes()
	log.Println("Starting web server on port 8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
	}

}
