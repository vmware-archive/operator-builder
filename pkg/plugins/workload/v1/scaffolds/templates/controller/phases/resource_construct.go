package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceConstruct{}

// ResourceConstruct scaffolds the construct phase methods.
type ResourceConstruct struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourceConstruct) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource_construct.go")

	f.TemplateBody = resourceConstructTemplate

	return nil
}

const resourceConstructTemplate = `{{ .Boilerplate }}

package phases

import (
	"{{ .Repo }}/apis/common"
)

// ConstructPhase.Execute executes the creation of resources in memory, prior to mutating or persisting them to the database.
func (phase *ConstructPhase) Execute(
	r common.ComponentReconciler,
	parentPhase *CreateResourcesPhase,
) (proceedToNextPhase bool, err error) {
	resources, err := r.GetResources(r.GetComponent())
	if err != nil {
		return false, err
	}

	// update the resources on the parent phase object
	setResources(parentPhase, resources)

	return true, nil
}
`