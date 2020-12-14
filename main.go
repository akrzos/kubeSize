package main

import (
	"github.com/akrzos/kubeSize/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cmd.Execute()
}
