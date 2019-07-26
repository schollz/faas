package utils

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/schollz/logger"
)

// GetStringInBetween returns empty string if no start or end string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func RunCommand(commands string, commandDuration ...time.Duration) (string, string, error) {
	tDuration := 100 * time.Hour
	if len(commandDuration) > 0 {
		tDuration = commandDuration[0]
	}

	log.Debugf("running %s", commands)
	command := strings.Fields(commands)
	cmd := exec.Command(command[0])
	if len(command) > 0 {
		cmd = exec.Command(command[0], command[1:]...)
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Start()
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(tDuration):
		if err := cmd.Process.Kill(); err != nil {
			log.Debug("failed to kill: ", err)
		}
		log.Debugf("%s killed as timeout reached", commands)
	case err := <-done:
		if err != nil {
			log.Debugf("err running %s: %s", commands, err.Error())
		}
	}
	return strings.TrimSpace(outb.String()), strings.TrimSpace(errb.String()), nil
}
