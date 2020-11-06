package main

import (
	"github.com/akrzos/k8sCube/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cmd.Execute()
}
