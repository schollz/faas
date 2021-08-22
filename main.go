package main

//go:generate cp -r pkg/gofaas/template .

import (
	"bytes"
	"encoding/base32"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/muesli/cache2go"
	"github.com/schollz/faas/pkg/gofaas"
	"github.com/schollz/faas/pkg/utils"
	log "github.com/schollz/logger"
)

type Server struct {
	HashToPort map[string]string
	sync.Mutex
}

func main() {
	var debug bool
	var port string
	flag.StringVar(&port, "port", "8090", "port to run")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()
	if debug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}

	os.Mkdir("images", os.ModePerm)
	s := new(Server)
	s.HashToPort = make(map[string]string)

	cache := cache2go.Cache("faasCache")

	log.Infof("running on port %s", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		urlPath := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
		log.Infof("%s?%s %s", r.URL.Path, r.URL.RawQuery, time.Since(timeStart))
		defer func() {
			log.Infof("%s?%s %s", r.URL.Path, r.URL.RawQuery, time.Since(timeStart))
		}()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			return
		}

		var body []byte
		var err error

		postBody, err := ioutil.ReadAll(r.Body)
		if err == nil {
			r.Body = ioutil.NopCloser(bytes.NewBuffer(postBody))
			urlPath += fmt.Sprintf("%s", postBody)
		}
		res, err := cache.Value(urlPath)
		if err != nil {
			body, err = s.handle(w, r)
			if err != nil {
				body = []byte(`{"success":false,"err":"` + err.Error() + `"}`)
			}
		} else {
			body = res.Data().([]byte)
			go func() {
				cache.Add(urlPath, 30*time.Minute, body)
			}()
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body)
	})
	http.ListenAndServe(":"+port, nil)
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) (body []byte, err error) {

	if r.Method == "OPTIONS" {
		w.Write([]byte("ok"))
		return
	}
	fString, ok := r.URL.Query()["func"]
	if !ok {
		err = fmt.Errorf("no func string")
		log.Error(err)
		return
	}
	funcString := strings.TrimSpace(strings.Split(fString[0], "(")[0])

	iString, ok := r.URL.Query()["import"]
	if !ok {
		err = fmt.Errorf("no import or url")
		log.Error(err)
		return
	}
	importPath := iString[0]

	id := getID(importPath, funcString)
	log.Debugf("funcString: [%s], importString: [%s]: %s", funcString, importPath, id)

	if !utils.Exists(path.Join("images", id+".tar.gz")) {
		log.Debugf("creating image for %s, %s()", importPath, funcString)
		err = gofaas.BuildContainer(importPath, funcString, id)
		if err != nil {
			log.Error(err)
			return
		}
	}

	// check if image is running
	stdout, stderr, err := utils.RunCommand("docker container ls --format '{{.Image}}'")
	log.Debugf("stdout: [%s]", stdout)
	log.Debugf("stderr: [%s]", stderr)
	if err != nil {
		log.Error(err)
		return
	}
	if stderr != "" {
		err = fmt.Errorf("%s", stderr)
		return
	}
	if !strings.Contains(stdout, id) {
		// run the image
		stdout, stderr, err = utils.RunCommand(fmt.Sprintf("docker load --input %s", path.Join("images", id+".tar.gz")))
		log.Debugf("stdout: [%s]", stdout)
		log.Debugf("stderr: [%s]", stderr)
		if err != nil {
			log.Error(err)
			return
		}
		if stderr != "" {
			err = fmt.Errorf("%s", stderr)
			return
		}
		stdout, stderr, err = utils.RunCommand(fmt.Sprintf("docker load --input %s", path.Join("images", id+".tar.gz")))
		log.Debugf("stdout: [%s]", stdout)
		log.Debugf("stderr: [%s]", stderr)
		if err != nil {
			log.Error(err)
			return
		}
		if stderr != "" {
			err = fmt.Errorf("%s", stderr)
			return
		}

		port := getOpenPort()
		log.Debugf("running on port %s", port)
		stdout, stderr, err = utils.RunCommand(fmt.Sprintf("docker run -d -t -p %s:8080 %s", port, id))
		log.Debugf("stdout: [%s]", stdout)
		log.Debugf("stderr: [%s]", stderr)
		if err != nil {
			log.Error(err)
			return
		}
		if stderr != "" {
			err = fmt.Errorf("%s", stderr)
			return
		}
	}

	// its running, get port
	stdout, stderr, err = utils.RunCommand("docker container ls")
	log.Debugf("stdout: [%s]", stdout)
	log.Debugf("stderr: [%s]", stderr)
	if err != nil {
		log.Error(err)
		return
	}
	if stderr != "" {
		err = fmt.Errorf("%s", stderr)
		return
	}
	portFound := ""
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, id) {
			portFound = utils.GetStringInBetween(line, "0.0.0.0:", "->")
			break
		}
	}
	if portFound == "" {
		err = fmt.Errorf("no port found")
		return
	}

	redirectURL := fmt.Sprintf("http://localhost:%s%s?%s", portFound, r.URL.Path, r.URL.RawQuery)
	log.Debugf("getting data from %s", redirectURL)
	var resp *http.Response
	if r.Method == "GET" {
		resp, err = http.Get(redirectURL)
	} else if r.Method == "POST" {
		resp, err = http.Post(redirectURL, "application/json", r.Body)
	} else {
		err = fmt.Errorf("not implemented")
	}
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func getID(importPath string, funcString string) (id string) {
	id = base32.StdEncoding.EncodeToString([]byte(importPath + " " + funcString))
	id = strings.Replace(id, "=", "", -1)
	id = "faas-" + strings.ToLower(id)
	return
}

func getOpenPort() (port string) {
	for i := 7000; i < 9000; i++ {
		port = strconv.Itoa(i)
		ln, err := net.Listen("tcp", ":"+port)
		if err == nil {
			ln.Close()
			break
		}
	}
	return
}
