package main

import (
	"fmt"
	"os"
)

const (
	functpl = `package main

{{ .FuncBody }}`
	maintpl = `package main

import (
	"flag"
	"fmt"
	"os"

	uselessruntime "github.com/damnever/useless/runtime"
)

func main() {
	laddr := flag.String("laddr", ":8080", "the listen address")
	flag.Parse()

	useless := uselessruntime.NewSupervisor("{{ .FuncName }}", {{ .FuncName }})
	defer useless.Close()
	if err := useless.Run(*laddr); err != nil {
		fmt.Fprintf(os.Stderr, "Launch function supervisor failed: %v", err)
		os.Exit(1)
	}
}`
)

func build(content, name, dockerReg string) {
	err := os.RemoveAll("./bin/func-main")
	assert(err == nil || os.IsNotExist(err), "rm -rf bin/func-main: %v", err)
	err = os.Mkdir("./bin/func-main", 0755)
	assert(err == nil || os.IsExist(err), "mkdir bin/func-main: %v", err)
	writeTemplate("./bin/func-main/main.go", maintpl, struct {
		FuncName string
	}{
		FuncName: name,
	})
	writeTemplate("./bin/func-main/func.go", functpl, struct {
		FuncBody string
	}{
		FuncBody: content,
	})
	execCmd("go build -o ./bin/function ./bin/func-main/", "GOOS=linux", "GOARCH=amd64", "GO111MODULE=on")
	imageName := imageName(name, dockerReg)
	execCmd(fmt.Sprintf("docker build -t %s -f ./docker/supervisor.Dockerfile --build-arg listen_addr=:80 .", imageName))
	execCmd(fmt.Sprintf("docker push %s", imageName))
}
