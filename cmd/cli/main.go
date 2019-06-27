package main

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

func main() {
	var (
		flagBuild      string
		flagCreate     string
		flagDelete     string
		flagDockerReg  string
		flagKubeConfig string
	)
	flag.StringVar(&flagBuild, "build", "", "build function image by <file-path>::<func-name>")
	flag.StringVar(&flagCreate, "create", "", "create and deploy function by <file-path>::<func-name>")
	flag.StringVar(&flagDelete, "delete", "", "delete function by meta name")
	flag.StringVar(&flagDockerReg, "docker-registry", "registry.cn-hangzhou.aliyuncs.com/useless", "docker registry")
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&flagKubeConfig, "kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&flagKubeConfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	switch {
	case flagBuild != "":
		content, name := readFunc(flagBuild)
		build(content, name, flagDockerReg)
	case flagCreate != "":
		content, name := readFunc(flagCreate)
		createFunction(content, name, flagDockerReg, flagKubeConfig)
	case flagDelete != "":
		deleteFunction(flagDelete, flagKubeConfig)
	default:
		assert(false, "None of -build/-create/-delete supplied.")
	}
}
