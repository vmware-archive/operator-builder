// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
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

	SpecFields        *workloadv1.APIFields
	IsComponent       bool
	ComponentResource *resource.Resource

	InitCommandName  string
	InitCommandDescr string
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

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr
	f.TemplateBody = cliCmdInitSubTemplate

	return nil
}

const cliCmdInitSubTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"
	"os"

	cmdinit "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/init"
)

const defaultManifest{{ .SubCmd.VarName }} = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
{{ .SpecFields.GenerateSampleSpec -}}
` + "`" + `

func Init{{ .Resource.Kind }}(i *cmdinit.InitSubCommand) error {
	outputStream := os.Stdout

	if _, err := outputStream.WriteString(defaultManifest); err != nil {
		return fmt.Errorf("failed to write to stdout, %w", err)
	}

	return nil
}
`
