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

	PackageName     string
	CreateFuncNames []string
	InitFuncNames   []string
	IsComponent     bool
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
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	"sigs.k8s.io/controller-runtime/pkg/client"
	
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

// Sample{{ .Resource.Kind }} returns the sample manifest for this custom resource.
func Sample{{ .Resource.Kind }}() string {
	return sample{{ .Resource.Kind }}
}

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{- if $.IsComponent }}
	*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }},
	{{ end -}}
) (client.Object, error){
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
	{{- if $.IsComponent }}
	*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }},
	{{ end -}}
) (client.Object, error){
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
