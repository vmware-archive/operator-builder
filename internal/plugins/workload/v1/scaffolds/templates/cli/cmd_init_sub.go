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

var _ machinery.Template = &CmdInitSub{}

// CmdInitSub scaffolds the companion CLI's init subcommand for the
// workload.  This where the actual init logic lives.
type CmdInitSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.RepositoryMixin

	RootCmd workloadv1.CliCommand
	SubCmd  workloadv1.CliCommand

	SpecFields                *workloadv1.APIFields
	IsStandalone, IsComponent bool
	ComponentResource         *resource.Resource

	InitCommandName  string
	InitCommandDescr string
	ManifestVarName  string
}

func (f *CmdInitSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Resource = f.ComponentResource
	}

	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"init",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)

	if f.IsStandalone {
		f.InitCommandName = initCommandName
		f.InitCommandDescr = initCommandDescr
	} else {
		f.InitCommandName = f.SubCmd.Name
		f.InitCommandDescr = f.SubCmd.Description
	}

	f.ManifestVarName = fmt.Sprintf("%sManifest%s", f.Resource.Version, f.Resource.Kind)

	f.TemplateBody = cliCmdInitSubTemplate

	return nil
}

const cliCmdInitSubTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdinit "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/init"
)

const {{ .ManifestVarName }} = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
{{ .SpecFields.GenerateSampleSpec -}}
` + "`" + `

// New{{ .Resource.Kind }}SubCommand creates a new command and adds it to its 
// parent command.
func New{{ .Resource.Kind }}SubCommand(parentCommand *cobra.Command) {
	initCmd := &cmdinit.InitSubCommand{
		Name:         "{{ .InitCommandName }}",
		Description:  "{{ .InitCommandDescr }}",
		InitFunc:     Init{{ .Resource.Kind }},
		SubCommandOf: parentCommand,
	}

	initCmd.Setup()
}

func Init{{ .Resource.Kind }}(i *cmdinit.InitSubCommand) error {
	outputStream := os.Stdout

	if _, err := outputStream.WriteString({{ .ManifestVarName }}); err != nil {
		return fmt.Errorf("failed to write to stdout, %w", err)
	}

	return nil
}
`
