// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var (
	_ machinery.Template = &CmdRoot{}
	_ machinery.Inserter = &CmdRootUpdater{}
)

// CmdRoot scaffolds the root command file for the companion CLI.
type CmdRoot struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	RootCmd workloadv1.CliCommand

	IsCollection bool
}

func (f *CmdRoot) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmd.Name, "commands", "root.go")

	f.TemplateBody = fmt.Sprintf(CmdRootTemplate,
		machinery.NewMarkerFor(f.Path, subcommandsImportsMarker),
		machinery.NewMarkerFor(f.Path, subcommandsGenerateMarker),
		machinery.NewMarkerFor(f.Path, subcommandsMarker),
	)

	return nil
}

// CmdRootUpdater updates root.go to run sub commands.
type CmdRootUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	RootCmdName string

	IsComponent, IsCollection, IsStandalone bool

	// Flags to indicate which parts need to be included when updating the file
	InitCommand, GenerateCommand, VersionCommand bool
}

// GetPath implements file.Builder interface.
func (f *CmdRootUpdater) GetPath() string {
	return filepath.Join("cmd", f.RootCmdName, "commands", "root.go")
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdRootUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const subcommandsImportsMarker = "operator-builder:subcommands:imports"
const subcommandsGenerateMarker = "operator-builder:subcommands:generate"
const subcommandsMarker = "operator-builder:subcommands"

// GetMarkers implements file.Inserter interface.
func (f *CmdRootUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), subcommandsImportsMarker),
		machinery.NewMarkerFor(f.GetPath(), subcommandsGenerateMarker),
		machinery.NewMarkerFor(f.GetPath(), subcommandsMarker),
	}
}

// Command Code Fragments.
const (
	initCommandCodeFragment = `// add the init subcommand
	c.newInitSubCommand(init%s.Init%s)

`
	generateCommandCodeFragment = `// add the generate subcommands
	generate%s.New%sSubCommand(parentCommand.Command)

`
	versionCommandCodeFragment = `// add the version subcommand
	c.newVersionSubCommand(version%s.Version%s)
	
`
)

// Import Code Fragments.
const (
	importSubCommandCodeFragment = `%s%s "%s"
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdRootUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate a command path for imports
	commandPath := fmt.Sprintf("%s/cmd/%s/commands", f.Repo, f.RootCmdName)

	// Generate subCommands and imports code fragments
	imports := make([]string, 0)
	subCommands := make([]string, 0)
	generateCommands := make([]string, 0)

	if f.InitCommand {
		imports = append(imports, fmt.Sprintf(importSubCommandCodeFragment,
			"init",
			f.Resource.Group,
			fmt.Sprintf("%s/init/%s", commandPath, f.Resource.Group)),
		)

		subCommands = append(subCommands, fmt.Sprintf(initCommandCodeFragment,
			f.Resource.Group,
			f.Resource.Kind),
		)
	}

	if f.GenerateCommand {
		imports = append(imports, fmt.Sprintf(importSubCommandCodeFragment,
			"generate",
			f.Resource.Group,
			fmt.Sprintf("%s/generate/%s", commandPath, f.Resource.Group)),
		)

		generateCommands = append(generateCommands, fmt.Sprintf(generateCommandCodeFragment,
			f.Resource.Group,
			f.Resource.Kind,
		),
		)
	}

	if f.VersionCommand {
		imports = append(imports, fmt.Sprintf(importSubCommandCodeFragment,
			"version",
			f.Resource.Group,
			fmt.Sprintf("%s/version/%s", commandPath, f.Resource.Group)),
		)

		subCommands = append(subCommands, fmt.Sprintf(versionCommandCodeFragment,
			f.Resource.Group,
			f.Resource.Kind),
		)
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsImportsMarker)] = imports
	}

	if len(subCommands) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsGenerateMarker)] = generateCommands
	}

	if len(subCommands) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsMarker)] = subCommands
	}

	return fragments
}

const CmdRootTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"

	// common imports for subcommands
	cmdinit "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/init"
	cmdgenerate "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/generate"
	cmdversion "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/version"

	// specific imports for workloads
	%s
)

// {{ .RootCmd.VarName }}Command represents the base command when called without any subcommands.
type {{ .RootCmd.VarName }}Command struct {
	*cobra.Command
}

// New{{ .RootCmd.VarName }}Command returns an instance of the {{ .RootCmd.VarName }}Command.
func New{{ .RootCmd.VarName }}Command() *{{ .RootCmd.VarName }}Command {
	c := &{{ .RootCmd.VarName }}Command{
		Command: &cobra.Command{
			Use:   "{{ .RootCmd.Name }}",
			Short: "{{ .RootCmd.Description }}",
			Long:  "{{ .RootCmd.Description }}",
		},
	}

	c.addSubCommands()

	return c
}

// Run represents the main entry point into the command
// This is called by main.main() to execute the root command.
func (c *{{ .RootCmd.VarName }}Command) Run() {
	cobra.CheckErr(c.Execute())
}

func (c *{{ .RootCmd.VarName }}Command) newInitSubCommand(initFunc cmdinit.InitFunc) {
	c.AddCommand(cmdinit.NewInitCommand(initFunc))
}

func (c *{{ .RootCmd.VarName }}Command) newGenerateSubCommand() {
	{{ if .IsCollection }}
	parentCommand := cmdgenerate.NewBaseGenerateSubCommand(c)
	{{ else }}
	// FIXME
	parentCommand := c
	_ = parentCommand
	{{ end }}

	%s
}

func (c *{{ .RootCmd.VarName }}Command) newVersionSubCommand(versionFunc cmdversion.VersionFunc) {
	c.AddCommand(cmdversion.NewVersionCommand(versionFunc))
}

// addSubCommands adds any additional subCommands to the root command.
func (c *{{ .RootCmd.VarName }}Command) addSubCommands() {
	c.newGenerateSubCommand()
	%s
}
`
