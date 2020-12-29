package main

import (
	"log"
	"net/http"

	"github.com/Hu13er/navaakache/cacheproxy"
)

func main() {
	log.Fatalln(
		http.ListenAndServe(":8000", cacheproxy.DefaultNavaakCache))
}
