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

	IsCollection bool

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
	"fmt"

	"github.com/spf13/cobra"
)

type InitFunc func(*InitSubCommand) error

type InitSubCommand struct {
	*cobra.Command

	// flags
	APIVersion string

	// options
	Name         string
	Description  string
	SubCommandOf *cobra.Command

	InitFunc InitFunc
}

{{ if .IsCollection }}
// NewBaseInitSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseInitSubCommand(parentCommand *cobra.Command) *InitSubCommand {
	initCmd := &InitSubCommand{
		Name:         "{{ .InitCommandName }}",
		Description:  "{{ .InitCommandDescr }}",
		SubCommandOf: parentCommand,
	}

	initCmd.Setup()

	return initCmd
}
{{ end }}

// Setup sets up this command to be used as a command.
func (i *InitSubCommand) Setup() {
	i.Command = &cobra.Command{
		Use:   i.Name,
		Short: i.Description,
		Long:  i.Description,
	}

	// run the initialize function if the function signature is set
	if i.InitFunc != nil {
		i.RunE = i.initialize
	}

	// always add the api-version flag
	i.Flags().StringVarP(
		&i.APIVersion,
		"api-version",
		"",
		"",
		"API Version of the workload to generate workload manifest for.",
	)

	// add this as a subcommand of another command if set
	if i.SubCommandOf != nil {
		i.SubCommandOf.AddCommand(i.Command)
	}
}

// GetParent is a convenience function written when the CLI code is scaffolded 
// to return the parent command and avoid scaffolding code with bad imports.
func GetParent(c interface{}) *cobra.Command {
	switch subcommand := c.(type) {
	case *InitSubCommand:
		return subcommand.Command
	case *cobra.Command:
		return subcommand
	}

	panic(fmt.Sprintf("subcommand is not proper type: %T", c))
}

// initialize creates sample workload manifests for a workload's custom resource.
func (i *InitSubCommand) initialize(cmd *cobra.Command, args []string) error {
	return i.InitFunc(i)
}
`
