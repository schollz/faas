package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/googleit"
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
	q := r.URL.Query().Get("q")
	var response []byte

	qstring, err := url.QueryUnescape(q)
	if err == nil {
		result, err := googleit.Search(qstring, googleit.Options{NumPages: 3, MustInclude: strings.Fields(qstring)})
		if err == nil {
			response, err = json.Marshal(result)
		}
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
