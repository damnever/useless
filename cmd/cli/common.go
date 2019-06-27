package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const (
	dirtyImageVersion = "latest"
)

func imageName(funcName, dockerReg string) string {
	return fmt.Sprintf("%s/%s:%s", dockerReg, strings.ToLower(funcName), dirtyImageVersion)
}

func assert(ok bool, format string, a ...interface{}) {
	if !ok {
		fmt.Fprintf(os.Stderr, format, a...)
		os.Exit(1)
	}
}

func readFunc(pathFunc string) (string, string) {
	parts := strings.SplitN(pathFunc, "::", 2)
	assert(len(parts) == 2, "format like this: <file-path>::<func-name>")
	path, name := parts[0], parts[1]
	content, err := ioutil.ReadFile(path)
	assert(err == nil, "read function failed: %v", err)
	newcontent := []byte{}
	for _, line := range bytes.Split(content, []byte{'\n'}) {
		if bytes.HasPrefix(bytes.TrimPrefix(line, []byte(" ")), []byte("package")) {
			continue
		}
		newcontent = append(newcontent, line...)
		newcontent = append(newcontent, '\n')
	}
	return string(newcontent), name
}

func writeTemplate(fpath, tplstr string, args interface{}) {
	tpl := template.Must(template.New("TODO").Parse(tplstr))
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, 0644)
	assert(err == nil, "open file: %v", err)
	defer f.Close()
	err = tpl.Execute(f, args)
	assert(err == nil, "generate code: %v", err)
}

func execCmd(command string, envs ...string) {
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = append(os.Environ(), envs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	assert(err == nil, "%s: %v", command, err)
}
