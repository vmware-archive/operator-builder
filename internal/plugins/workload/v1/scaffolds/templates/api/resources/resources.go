// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &Resources{}

// Types scaffolds child resource creation functions.
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	RootCmdName     string
	PackageName     string
	CreateFuncNames []string
	InitFuncNames   []string
	IsComponent     bool
	IsStandalone    bool
	IsCollection    bool
	Collection      *workloadv1.WorkloadCollection
	SpecFields      *workloadv1.APIFields
}

func (f *Resources) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.PackageName,
		"resources.go",
	)

	f.TemplateBody = resourcesTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint:lll
const resourcesTemplate = `{{ .Boilerplate }}

package {{ .PackageName }}

import (
	{{ if ne .RootCmdName "" }}"fmt"{{ end }}

	{{ if ne .RootCmdName "" }}"sigs.k8s.io/yaml"{{ end }}
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	
	{{ if ne .RootCmdName "" }}cmdutils "{{ .Repo }}/cmd/{{ .RootCmdName }}/commands/utils"{{ end }}
	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .IsComponent }}
	{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Collection.Spec.API.Group }}/{{ .Collection.Spec.API.Version }}"
	{{ end -}}
)

const sample{{ .Resource.Kind }} = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
{{ .SpecFields.GenerateSampleSpec -}}` + "`" + `

// Sample returns the sample manifest for this custom resource.
func Sample() string {
	return sample{{ .Resource.Kind }}
}

// Generate returns the child resources that are associated with this workload given 
// appropriate structured inputs.
{{ if .IsComponent -}}
func Generate(
	workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}, 
	collection {{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }},
) ([]client.Object, error) {
{{ else if .IsCollection -}}
func Generate(collection {{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }}) ([]client.Object, error) {
{{ else -}}
func Generate(workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}) ([]client.Object, error) {
{{ end -}}
	resourceObjects := make([]client.Object, len(CreateFuncs))

	for i, f := range CreateFuncs {
		{{ if .IsComponent -}}
		resource, err := f(&workload, &collection)
		{{ else if .IsCollection -}}
		resource, err := f(&collection)
		{{ else -}}
		resource, err := f(&workload)
		{{ end }}
		if err != nil {
			return nil, err
		}

		resourceObjects[i] = resource
	}

	return resourceObjects, nil
}

{{ if ne .RootCmdName "" }}
// GenerateForCLI returns the child resources that are associated with this workload given
// appropriate YAML manifest files.
func GenerateForCLI(
	{{- if or (.IsStandalone) (.IsComponent) }}workloadFile []byte,{{ end -}}
	{{- if or (.IsComponent) (.IsCollection) }}collectionFile []byte,{{ end -}}
) ([]client.Object, error) {
	{{- if or (.IsStandalone) (.IsComponent) }}
	var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}
	if err := yaml.Unmarshal(workloadFile, &workload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into workload, %w", err)
	}

	if err := cmdutils.ValidateWorkload(&workload); err != nil {
		return nil, fmt.Errorf("error validating workload yaml, %w", err)
	}
	{{ end }}

	{{- if or (.IsComponent) (.IsCollection) }}
	var collection {{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }}
	if err := yaml.Unmarshal(collectionFile, &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into collection, %w", err)
	}

	if err := cmdutils.ValidateWorkload(&collection); err != nil {
		return nil, fmt.Errorf("error validating collection yaml, %w", err)
	}
	{{ end }}

	{{ if .IsComponent }}
	return Generate(workload, collection)
	{{ else if .IsCollection }}
	return Generate(collection)
	{{ else }}
	return Generate(workload)
	{{ end -}}
}
{{ end }}

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{ if $.IsComponent -}}
	*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }},
	{{ end -}}
) (client.Object, error) {
	{{ range .CreateFuncNames }}
		{{- . -}},
	{{ end }}
}

// InitFuncs is an array of functions that are called prior to starting the controller manager.  This is
// necessary in instances which the controller needs to "own" objects which depend on resources to
// pre-exist in the cluster. A common use case for this is the need to own a custom resource.
// If the controller needs to own a custom resource type, the CRD that defines it must
// first exist. In this case, the InitFunc will create the CRD so that the controller
// can own custom resources of that type.  Without the InitFunc the controller will
// crash loop because when it tries to own a non-existent resource type during manager
// setup, it will fail.
var InitFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{ if $.IsComponent -}}
	*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }},
	{{ end -}}
) (client.Object, error) {
	{{ range .InitFuncNames }}
		{{- . -}},
	{{ end }}
}

{{ if $.IsComponent -}}
func ConvertWorkload(component, collection workload.Workload) (
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }},
	error,
) {
{{- else }}
func ConvertWorkload(component workload.Workload) (*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}, error) {
{{- end }}
	p, ok := component.(	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }})
	if !ok {
		{{- if $.IsComponent }}
		return nil, nil, {{ .Resource.ImportAlias }}.ErrUnableToConvert{{ .Resource.Kind }}
	}

	c, ok := collection.(*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }})
	if !ok {
		return nil, nil, {{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.ErrUnableToConvert{{ .Collection.Spec.API.Kind }}
	}

	return p, c, nil
{{- else }}
		return nil, {{ .Resource.ImportAlias }}.ErrUnableToConvert{{ .Resource.Kind }}
  }

	return p, nil
{{- end }}
}
`
