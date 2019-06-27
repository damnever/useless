package main

import (
	"fmt"
	"strings"

	uselessv1 "github.com/damnever/useless/pkg/apis/useless/v1"
	"github.com/damnever/useless/pkg/generated/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace = "useless"
)

func createFunction(content, name, dockerReg, kubeConfig string) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	assert(err == nil, "build config failed: %v", err)

	defaultReplicas := int32(1)
	funcclientset, err := versioned.NewForConfig(config)
	assert(err == nil, "create clientset failed: %v", err)
	name = strings.ToLower(name) // XXX(damnever): to lower, fuck..
	function, err := funcclientset.UselessV1().Functions(defaultNamespace).Create(
		&uselessv1.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: defaultNamespace,
			},
			Spec: uselessv1.FunctionSpec{
				FuncName:    name,
				FuncContent: content,
				Image:       imageName(name, dockerReg),
				Replicas:    &defaultReplicas,
			},
		},
	)
	assert(err == nil, "create function failed: %v", err)
	fmt.Printf("Function create, name: %s - %s\n", function.GetName(), function.Spec.FuncName)
}

func deleteFunction(name, kubeConfig string) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	assert(err == nil, "build config failed: %v", err)

	funcclientset, err := versioned.NewForConfig(config)
	assert(err == nil, "create clientset failed: %v", err)
	err = funcclientset.UselessV1().Functions(defaultNamespace).Delete(name, nil)
	assert(err == nil, "create function failed: %v", err)
}
