// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &CmdGenerateSub{}

// CmdGenerateSub scaffolds the companion CLI's generate subcommand for the
// workload.  This where the actual generate logic lives.
type CmdGenerateSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	PackageName string

	RootCmd workloadv1.CliCommand
	SubCmd  workloadv1.CliCommand

	IsComponent, IsCollection, IsStandalone bool

	ComponentResource *resource.Resource
	Collection        *workloadv1.WorkloadCollection

	GenerateCommandName       string
	GenerateCommandDescr      string
	UseCollectionManifestFlag bool
	UseWorkloadManifestFlag   bool
}

func (f *CmdGenerateSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Resource = f.ComponentResource
	}

	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"generate",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)

	// if we have a standalone simply use the default command name and description
	// for generate since the 'generate' command will be the last in the chain,
	// otherwise we will use the requested subcommand name
	if f.IsStandalone {
		f.GenerateCommandName = generateCommandName
		f.GenerateCommandDescr = generateCommandDescr
	} else {
		f.GenerateCommandName = f.SubCmd.Name
		f.GenerateCommandDescr = f.SubCmd.Description
	}

	// use the collection flag for non-standalone use cases
	if !f.IsStandalone {
		f.UseCollectionManifestFlag = true
	}

	// use the workload manifest flag for non-collection use cases
	if !f.IsCollection {
		f.UseWorkloadManifestFlag = true
	}

	f.TemplateBody = cliCmdGenerateSubTemplate

	return nil
}

//nolint: lll
const cliCmdGenerateSubTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	cmdgenerate "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/generate"
	cmdutils "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/utils"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	"{{ .Resource.Path }}/{{ .PackageName }}"
	{{- if .IsComponent }}
	{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Collection.Spec.API.Group }}/{{ .Collection.Spec.API.Version }}"
	{{ end -}}
)

// New{{ .Resource.Kind }}SubCommand creates a new command and adds it to its 
// parent command.
func New{{ .Resource.Kind }}SubCommand(parentCommand *cobra.Command) {
	generateCmd := &cmdgenerate.GenerateSubCommand{
		Name:                  "{{ .GenerateCommandName }}",
		Description:           "{{ .GenerateCommandDescr }}",
		UseCollectionManifest: {{ .UseCollectionManifestFlag }},
		UseWorkloadManifest:   {{ .UseWorkloadManifestFlag }},
		SubCommandOf:          parentCommand,
		GenerateFunc:          Generate{{ .Resource.Kind }},
	}

	generateCmd.Setup()
}

// Generate{{ .Resource.Kind }} runs the logic to generate child resources for a
// {{ .Resource.Kind }} workload.
func Generate{{ .Resource.Kind }}(g *cmdgenerate.GenerateSubCommand) error {
	{{- if and (.IsComponent) (not .IsCollection) }}
	// component workload
	wkFilename, _ := filepath.Abs(g.WorkloadManifest)

	wkYamlFile, err := os.ReadFile(wkFilename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", wkFilename, err)
	}

	var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}

	err = yaml.Unmarshal(wkYamlFile, &workload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", wkFilename, err)
	}

	err = cmdutils.ValidateWorkload(&workload)
	if err != nil {
		return fmt.Errorf("error validating yaml %s, %w", wkFilename, err)
	}

	{{ end -}}
	{{- if .IsComponent }}
	// workload collection
	colFilename, _ := filepath.Abs(g.CollectionManifest)

	colYamlFile, err := os.ReadFile(colFilename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", colFilename, err)
	}

	var collection {{ $.Collection.Spec.API.Group }}{{ $.Collection.Spec.API.Version }}.{{ $.Collection.Spec.API.Kind }}

	err = yaml.Unmarshal(colYamlFile, &collection)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", colFilename, err)
	}

	err = cmdutils.ValidateWorkload(&collection)
	if err != nil {
		return fmt.Errorf("error validating yaml %s, %w", colFilename, err)
	}

	resourceObjects := make([]client.Object, len({{ .PackageName }}.CreateFuncs))

	for i, f := range {{ .PackageName }}.CreateFuncs {
		{{- if .IsCollection }}
		resource, err := f(&collection)
		{{- else }}
		resource, err := f(&workload, &collection)
		{{- end }}
		if err != nil {
			return err
		}

		resourceObjects[i] = resource
	}
	{{ else }}
	filename, _ := filepath.Abs(g.WorkloadManifest)

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", filename, err)
	}

	var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}

	err = yaml.Unmarshal(yamlFile, &workload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", filename, err)
	}

	err = cmdutils.ValidateWorkload(&workload)
	if err != nil {
		return fmt.Errorf("error validating yaml %s, %w", filename, err)
	}

	resourceObjects := make([]client.Object, len({{ .PackageName }}.CreateFuncs))

	for i, f := range {{ .PackageName }}.CreateFuncs {
		resource, err := f(&workload)
		if err != nil {
			return err
		}

		resourceObjects[i] = resource
	}
	{{ end }}

	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	outputStream := os.Stdout

	for _, o := range resourceObjects {
		if _, err := outputStream.WriteString("---\n"); err != nil {
			return fmt.Errorf("failed to write output, %w", err)
		}

		if err := e.Encode(o, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output, %w", err)
		}
	}

	return nil
}
`
