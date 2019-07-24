package main

import (
	"flag"
	"net/http"

	log "github.com/schollz/logger"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()
	if debug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}
	log.Infof("running on port %s", "8080")
	//http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

