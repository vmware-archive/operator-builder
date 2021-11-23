// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

const (
	generateCommandName  = "generate"
	generateCommandDescr = "Generate child resource manifests from a workload's custom resource"
)

var _ machinery.Template = &CmdGenerate{}

// CmdGenerate scaffolds the companion CLI's generate subcommand for
// component workloads.  The generate logic will live in the workload's
// subcommand to this command; see cmd_generate_sub.go.
type CmdGenerate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	RootCmdName string

	IsCollection bool

	GenerateCommandName  string
	GenerateCommandDescr string
}

func (f *CmdGenerate) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmdName, "commands", "generate", "generate.go")
	f.TemplateBody = cliCmdGenerateTemplate

	f.GenerateCommandName = generateCommandName
	f.GenerateCommandDescr = generateCommandDescr

	return nil
}

const cliCmdGenerateTemplate = `{{ .Boilerplate }}

package generate

import (
	"github.com/spf13/cobra"
)

type GenerateFunc func(*GenerateSubCommand) error

type GenerateSubCommand struct {
	*cobra.Command

	// flags
	WorkloadManifest   string
	CollectionManifest string
	APIVersion         string

	// options
	Name                  string
	Description           string
	UseCollectionManifest bool
	UseWorkloadManifest   bool
	SubCommandOf          *cobra.Command

	// execution
	GenerateFunc GenerateFunc
}

{{ if .IsCollection }}
// NewBaseGenerateSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseGenerateSubCommand(parentCommand *cobra.Command) *GenerateSubCommand {
	generateCmd := &cmdgenerate.GenerateSubCommand{
		Name:                  "{{ .GenerateCommandName }}",
		Description:           "{{ .GenerateCommandDescr }}",
		UseCollectionManifest: false,
		UseWorkloadManifest:   false,
		SubCommandOf:          parentCommand,
	}

	generateCmd.Setup()

	return generateCmd
}
{{ end }}

// Setup sets up this command to be used as a command.
func (g *GenerateSubCommand) Setup() {
	g.Command = &cobra.Command{
		Use:   g.Name,
		Short: g.Description,
		Long:  g.Description,
	}

	// run the generate function if the function signature is set
	if g.GenerateFunc != nil {
		g.RunE = g.generate
	}

	// add workload-manifest flag if this subcommand requests it
	if g.UseWorkloadManifest {
		g.Flags().StringVarP(
			&g.WorkloadManifest,
			"workload-manifest",
			"w",
			"",
			"Filepath to the workload manifest to generate child resources for.",
		)

		if err := g.MarkFlagRequired("workload-manifest"); err != nil {
			panic(err)
		}
	}

	// add collection-manifest flag if this subcommand requests it
	if g.UseCollectionManifest {
		g.Command.Flags().StringVarP(
			&g.CollectionManifest,
			"collection-manifest",
			"c",
			"",
			"Filepath to the workload collection manifest.",
		)

		if err := g.MarkFlagRequired("collection-manifest"); err != nil {
			panic(err)
		}
	}

	// always add the api-version flag
	g.Flags().StringVarP(
		&g.APIVersion,
		"api-version",
		"",
		"",
		"API Version of the workload to generate child resources for.",
	)

	// add this as a subcommand of another command if set
	if g.SubCommandOf != nil {
		g.SubCommandOf.AddCommand(g.Command)
	}
}

// generate creates child resource manifests from a workload's custom resource.
func (g *GenerateSubCommand) generate(cmd *cobra.Command, args []string) error {
	return g.GenerateFunc(g)
}
`
