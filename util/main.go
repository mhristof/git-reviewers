package util

import (
	"bytes"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Eval(command string) []string {
	cmd := exec.Command("/bin/bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr := stdout.String()
	if err != nil {
		log.WithFields(log.Fields{
			"outStr":  outStr,
			"command": command,
		}).Error("failed command")
	}

	return strings.Split(string(outStr), "\n")
}

func Keys(in map[string]bool) []string {
	var ret []string
	for key, _ := range in {
		ret = append(ret, key)
	}

	return ret
}

func Uniq(in []string) []string {
	m := map[string]bool{}

	for _, item := range in {
		m[item] = true
	}

	return Keys(m)
}
