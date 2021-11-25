// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var (
	_ machinery.Template = &CmdVersionSub{}
	_ machinery.Inserter = &CmdVersionSubUpdater{}
)

// CmdVersionSub scaffolds the root command file for the companion CLI.
type CmdVersionSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.RepositoryMixin

	RootCmd workloadv1.CliCommand
	SubCmd  workloadv1.CliCommand

	// VersionCommandName is the version sub command
	VersionCommandName  string
	VersionCommandDescr string

	// Variable Names
	APIVersionsVarName string

	IsComponent       bool
	IsStandalone      bool
	ComponentResource *resource.Resource
}

func (f *CmdVersionSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Resource = f.ComponentResource
	}

	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"version",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)

	f.VersionCommandName = versionCommandName
	f.VersionCommandDescr = versionCommandDescr

	if f.IsStandalone {
		f.APIVersionsVarName = "apiVersions"
	} else {
		f.APIVersionsVarName = fmt.Sprintf("apiVersions%s", f.Resource.Kind)
	}

	f.TemplateBody = fmt.Sprintf(
		cmdVersionSubHeader,
		machinery.NewMarkerFor(f.Path, apiVersionsMarker),
		cmdVersionSubBody,
	)

	return nil
}

// CmdVersionSubUpdater updates root.go to run sub commands.
type CmdVersionSubUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	RootCmd workloadv1.CliCommand
	SubCmd  workloadv1.CliCommand

	IsComponent bool
}

// GetPath implements file.Builder interface.
func (f *CmdVersionSubUpdater) GetPath() string {
	return f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"version",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdVersionSubUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const apiVersionsMarker = "operator-builder:apiversions"

// GetMarkers implements file.Inserter interface.
func (f *CmdVersionSubUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), apiVersionsMarker),
	}
}

// Code Fragments.
const (
	versionCodeFragment = `"%s",
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdVersionSubUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate subCommands code fragments
	apiVersions := make([]string, 0)
	apiVersions = append(apiVersions, fmt.Sprintf(versionCodeFragment, f.Resource.Version))

	// Only store code fragments in the map if the slices are non-empty
	if len(apiVersions) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), apiVersionsMarker)] = apiVersions
	}

	return fragments
}

const (
	cmdVersionSubHeader = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"github.com/spf13/cobra"
	
	cmdversion "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/version"
)

var {{ .APIVersionsVarName }} = []string{
	%s
}

%s
`
	cmdVersionSubBody = `
// New{{ .Resource.Kind }}SubCommand creates a new command and adds it to its 
// parent command.
func New{{ .Resource.Kind }}SubCommand(parentCommand *cobra.Command) {
	versionCmd := &cmdversion.VersionSubCommand{
		{{- if .IsStandalone }}
		Name:         "{{ .VersionCommandName }}",
		Description:  "{{ .VersionCommandDescr }}",
		{{ else }}
		Name:         "{{ .SubCmd.Name }}",
		Description:  "{{ .SubCmd.Description }}",
		{{- end -}}
		VersionFunc:  Version{{ .Resource.Kind }},
		SubCommandOf: parentCommand,
	}

	versionCmd.Setup()
}

func Version{{ .Resource.Kind }}(v *cmdversion.VersionSubCommand) error {
	versionInfo := cmdversion.VersionInfo{
		CLIVersion:  cmdversion.CliVersion,
		APIVersions: {{ .APIVersionsVarName }},
	}

	return versionInfo.Display()
}
`
)
