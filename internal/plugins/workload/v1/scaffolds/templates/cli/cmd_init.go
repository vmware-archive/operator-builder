// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

const (
	initCommandName  = "init"
	initCommandDescr = "Write a sample custom resource manifest for a workload to standard out"
)

var _ machinery.Template = &CmdInit{}

// CmdInit scaffolds the companion CLI's init subcommand for
// component workloads.  The init logic will live in the workload's
// subcommand to this command; see cmd_init_sub.go.
type CmdInit struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	RootCmdName string

	InitCommandName  string
	InitCommandDescr string
}

func (f *CmdInit) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmdName, "commands", "init", "init.go")
	f.TemplateBody = cliCmdInitTemplate

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	return nil
}

const cliCmdInitTemplate = `{{ .Boilerplate }}

package init

import (
	"github.com/spf13/cobra"
)

type InitFunc func(*InitSubCommand) error

type InitSubCommand struct {
	*cobra.Command

	APIVersion string

	initFunc InitFunc
}

// NewInitCommand creates a new instance of the init subcommand.
func NewInitCommand(initFunc InitFunc) *cobra.Command {
	i := &InitSubCommand{
		initFunc: initFunc,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Write a sample custom resource manifest for a workload to standard out",
		Long:  "Write a sample custom resource manifest for a workload to standard out",
		RunE:  i.initialize,
	}

	initCmd.Flags().StringVarP(
		&i.APIVersion,
		"api-version",
		"",
		"",
		"API Version of the workload to generate workload manifest for.",
	)

	return initCmd
}

// initialize creates sample workload manifests for a workload's custom resource.
func (i *InitSubCommand) initialize(cmd *cobra.Command, args []string) error {
	return i.initFunc(i)
}
`
