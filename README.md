## Useless

A simple useless 'FaaS' framework build on top of Kubernetes, since it is useless, hard coded things were everywhere.

Just for [learning](), play with it:
```Bash
# If you want to run ./hack/update-codegen.sh, you must:
#  - go get -u k8s.io/code-generator/cmd/...
#  - put the current project under /path/to/github.com/damnever/, that is a known issue for code-generator
git clone git@github.com:damnever/useless.git && cd useless

# NOTE: minikube may fail to set up the cluster if --memory too small(e.g. 1024)..
kubectl create namespace useless
kubectl config set-context --current --namespace=useless
# Create CRD
kubectl create -f ./artifacts/function-definition.yaml
# Create controller to deal with CRD
kubectl create -f ./artifacts/controller-deployment.yaml

make build-cli
# ./bin/useless-cli -build ./artifacts/what_the_commits.go::WhatTheCommits  # build and push function image
./bin/useless-cli -create ./artifacts/what_the_commits.go::WhatTheCommits

kubectl get services
# Ingress maybe a good choice, anyway..
kubectl port-forward service/whatthecommits 8080:80
curl -H "Content-Type: application/json" -X POST -d '{"input":"{\"count\":3}"}' http://localhost:8080
```
