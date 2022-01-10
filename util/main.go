package util

import (
	"bytes"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Eval Run a system command and return the stdout lines.
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

	return strings.Split(outStr, "\n")
}

// Keys Returns a list of keys for the given map.
func Keys(in map[string]struct{}) []string {
	var ret []string
	for key := range in {
		ret = append(ret, key)
	}

	return ret
}

// Uniq Return a uniq list of items in the input list.
func Uniq(in []string) []string {
	m := map[string]struct{}{}

	for _, item := range in {
		m[item] = struct{}{}
	}

	return Keys(m)
}

func Subtract(haystack []string, remove []string) []string {
	rem := map[string]bool{}

	for _, k := range remove {
		rem[k] = true
	}

	var diff []string

	for _, item := range haystack {
		if _, ok := rem[item]; ok {
			continue
		}

		diff = append(diff, item)
	}

	return diff
}

func Merge(a, b map[string]struct{}) map[string]struct{} {
	for k, v := range b {
		a[k] = v
	}

	return a
}
