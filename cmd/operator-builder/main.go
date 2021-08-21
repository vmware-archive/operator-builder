package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/cli"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1"
)

func main() {
	command, err := cli.NewKubebuilderCLI(workloadv1.CommandName)
	if err != nil {
		log.Fatal(err)
	}

	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}
