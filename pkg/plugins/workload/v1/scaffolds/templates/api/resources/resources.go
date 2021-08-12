package resources

import (
	"path/filepath"
	"text/template"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
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
	SpecFields      *[]workloadv1.APISpecField
	IsComponent     bool
	Collection      *workloadv1.WorkloadCollection
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

func (f *Resources) GetFuncMap() template.FuncMap {
	funcMap := machinery.DefaultFuncMap()
	funcMap["quotestr"] = func(value string) string {
		if string(value[0]) != `"` {
			value = `"` + value
		}

		if string(value[len(value)-1]) != `"` {
			value += `"`
		}

		return value
	}

	return funcMap
}

//nolint:lll
const resourcesTemplate = `{{ .Boilerplate }}

package {{ .PackageName }}

import (
	"fmt"
	"bytes"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .IsComponent }}
	{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }} "{{ .Repo }}/apis/{{ .Collection.Spec.APIGroup }}/{{ .Collection.Spec.APIVersion }}"
	{{ end -}}
)

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{- if $.IsComponent }}
	*{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }},
	{{ end -}}
) (metav1.Object, error){
	{{ range .CreateFuncNames }}
		{{- . -}},
	{{ end }}
}

// InitFuncs is an array of functions that are called prior to starting the controller manager.  This is
// necessary in instances which the controller needs to "own" objects which depend on resources to
// pre-exist in the cluster. A common use case for this is a CustomResourceDefinition and a resource
// created as part of that CRD.  In this case, we cannot create and own the CRD and its affiliated resource
// as part of the reconciliation loop.  Because the CRD does not exist, the API has not been extended
// properly for the controller to own the affiliated resource, and it will fail with an error.
var InitFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{- if $.IsComponent }}
	*{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }},
	{{ end -}}
) (metav1.Object, error){
	{{ range .InitFuncNames }}
		{{- . -}},
	{{ end }}
}

// runTemplate renders a template for a child object to the custom resource.
func runTemplate(templateName, templateValue string, data *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	funcMap template.FuncMap) (string, error) {
	t, err := template.New(templateName).Funcs(funcMap).Parse(templateValue)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %v", templateName, err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, &data); err != nil {
		return "", fmt.Errorf("error rendering template %s: %v", templateName, err)
	}

	return b.String(), nil
}

{{ range .SpecFields }}
{{ if .DefaultVal }}
{{ if eq .DataType "string" }}
const {{ .ManifestFieldName }}Default = {{ .DefaultVal | quotestr }}
{{ else }}
const {{ .ManifestFieldName }}Default = {{ .DefaultVal }}
{{ end }}

func default{{ .FieldName }}(value {{ .DataType }}) {{ .DataType }} {
	if value == {{ .ZeroVal }} {
		return {{ .ManifestFieldName }}Default
	}

	return value
}
{{ end }}
{{ end }}
`
