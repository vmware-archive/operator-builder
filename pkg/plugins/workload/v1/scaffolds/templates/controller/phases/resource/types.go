package resource

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Types{}

// Types scaffolds the phase interfaces and types.
type Types struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Types) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource", "types.go")

	f.TemplateBody = typesTemplate

	return nil
}

const typesTemplate = `{{ .Boilerplate }}

package resource

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// ResourcePhase defines the specific phase of reconcilication associated with creating resources.
type ResourcePhase interface {
	Execute(common.ComponentResource, common.ResourceCondition) (ctrl.Result, bool, error)
}

// Below are the phase types which satisfy the ResourcePhase interface.
type Persist struct{}
type Wait struct{}
`
