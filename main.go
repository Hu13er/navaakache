package main

import (
	"log"
	"net/http"

	"github.com/Hu13er/navaakache/proxy"
)

func main() {
	log.Fatalln(
		http.ListenAndServe(":8000", proxy.DefaultNavaakCache))
}
