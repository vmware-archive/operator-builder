// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
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

	// RootCmd is the root command for the companion CLI
	RootCmd        string
	RootCmdVarName string

	// SubCmdName is the sub command for the component
	SubCmdName     string
	SubCmdDescr    string
	SubCmdVarName  string
	SubCmdFileName string

	// VersionCommandName is the version sub command
	VersionCommandName  string
	VersionCommandDescr string

	IsComponent       bool
	ComponentResource *resource.Resource
}

func (f *CmdVersionSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Resource = f.ComponentResource
	}

	f.Path = getFilePath(f.IsComponent, f.RootCmd, f.SubCmdFileName)
	f.VersionCommandName = versionCommandName
	f.VersionCommandDescr = versionCommandDescr

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

	RootCmd        string
	SubCmdFileName string

	IsComponent bool
}

// GetPath implements file.Builder interface.
func (f *CmdVersionSubUpdater) GetPath() string {
	return getFilePath(f.IsComponent, f.RootCmd, f.SubCmdFileName)
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

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var default{{ .SubCmdVarName }} = "dev"
var apiVersions{{ .SubCmdVarName }} = []string{
	%s
}

%s
`
	cmdVersionSubBody = `
{{ if not .IsComponent -}}
// newVersionCommand creates a new instance of the version subcommand.
func (c *{{ .RootCmdVarName }}Command) newVersionCommand() {
{{- else }}
// newVersion{{ .SubCmdVarName }}Command creates a new instance of the version{{ .SubCmdVarName }} subcommand.
func (v *versionCommand) newVersion{{ .SubCmdVarName }}Command() {
{{- end }}
	version{{ .SubCmdVarName }}Cmd := &cobra.Command{
		{{ if .IsComponent -}}
		Use:   "{{ .SubCmdName }}",
		Short: "{{ .SubCmdDescr }}",
		Long:  "{{ .SubCmdDescr }}",
		{{- else -}}
		Use:   "{{ .VersionCommandName }}",
		Short: "{{ .VersionCommandDescr }}",
		Long:  "{{ .VersionCommandDescr }}",
		{{- end }}
		RunE: func(cmd *cobra.Command, args []string) error {
			versionInfo := struct {
				CLIVersion  string   ` + "`" + `json:"cliVersion"` + "`" + `
				APIVersions []string ` + "`" + `json:"apiVersions"` + "`" + `
			}{
				CLIVersion:  default{{ .SubCmdVarName }},
				APIVersions: apiVersions{{ .SubCmdVarName }},
			}

			output, err := json.Marshal(versionInfo)
			if err != nil {
				return fmt.Errorf("failed to determine versionInfo, %s", err)
			}

			outputStream := os.Stdout

			if _, err := outputStream.WriteString(fmt.Sprintln(string(output))); err != nil {
				return fmt.Errorf("failed to write to stdout, %s", err)
			}

			return nil
		},
	}

	{{ if .IsComponent -}}
	v.AddCommand(version{{ .SubCmdVarName }}Cmd)
	{{- else -}}
	c.AddCommand(version{{ .SubCmdVarName }}Cmd)
	{{- end -}}
}
`
)

func getFilePath(isComponent bool, rootCmdName, subCmdFileName string) string {
	if isComponent {
		return filepath.Join(
			"cmd", rootCmdName, "commands",
			fmt.Sprintf("%s_version.go", subCmdFileName),
		)
	}

	return filepath.Join("cmd", rootCmdName, "commands", "version.go")
}
