package controller

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
	f.Path = filepath.Join("controllers", "phases", "controller", "types.go")

	f.TemplateBody = typesTemplate

	return nil
}

const typesTemplate = `{{ .Boilerplate }}

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// Phase defines a phase of the reconciliation process.
type Phase interface {
	Execute(common.ComponentReconciler) (bool, error)
	DefaultRequeue() ctrl.Result
}

// Below are the phase types which satisfy the Phase interface.
type Dependency struct{}
type PreFlight struct{}
type CreateResources struct{}
type CheckReady struct{}
type Complete struct{}
`
