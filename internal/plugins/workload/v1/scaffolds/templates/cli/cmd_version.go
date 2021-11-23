// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

const (
	versionCommandName  = "version"
	versionCommandDescr = "Display the version information"
)

var _ machinery.Template = &CmdVersion{}

// CmdVersion scaffolds the companion CLI's version subcommand for
// component workloads.  The version logic will live in the workload's
// subcommand to this command; see cmd_version_sub.go.
type CmdVersion struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	RootCmd        string
	RootCmdVarName string

	VersionCommandName  string
	VersionCommandDescr string

	SubCommands *[]workloadv1.CliCommand
}

func (f *CmdVersion) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmd, "commands", "version.go")

	f.VersionCommandName = versionCommandName
	f.VersionCommandDescr = versionCommandDescr

	f.TemplateBody = cliCmdVersionTemplate

	return nil
}

const cliCmdVersionTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"
)

type versionCommand struct {
	*cobra.Command
}

// newVersionCommand creates a new instance of the version subcommand.
func (c *{{ .RootCmdVarName }}Command) newVersionCommand() {
	versionCmd := &versionCommand{}

	versionCmd.Command = &cobra.Command{
		Use:   "{{ .VersionCommandName }}",
		Short: "{{ .VersionCommandDescr }}",
		Long:  "{{ .VersionCommandDescr }}",
	}

	versionCmd.addCommands()
	c.AddCommand(versionCmd.Command)
}

func (v *versionCommand) addCommands() {
	{{- range $cmd := .SubCommands }}
	v.newVersion{{ $cmd.VarName }}Command()
	{{- end }}
}
`
