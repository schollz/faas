package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"strconv"
	"time"

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
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()
	defer func() {
		log.Infof("%s?%s %s", r.URL.Path, r.URL.RawQuery, time.Since(timeStart))
	}()

	var response []byte
	var err error
	if r.Method == "POST" {
		response, err = handlePost(w, r)
	} else if r.Method == "GET" {

	} else if r.Method == "OPTION" {
		response = []byte("ok")
	}

	if err != nil {
		res := struct {
			Message string `json:"message"`
			Success bool   `json:"success"`
		}{
			err.Error(),
			false,
		}
		response, _ = json.Marshal(res)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "text/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.Write(response)
}

func handlePost(w http.ResponseWriter, r *http.Request) (reponse []byte, err error) {
	// decoder := json.NewDecoder(r.Body)
	// err = decoder.Decode(&t)
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }

	return
}
