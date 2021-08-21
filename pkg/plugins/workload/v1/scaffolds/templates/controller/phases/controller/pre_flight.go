package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PreFlight{}

// PreFlight scaffolds the pre-flight phase methods.
type PreFlight struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *PreFlight) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "controller", "pre_flight.go")

	f.TemplateBody = preFlightTemplate

	return nil
}

const preFlightTemplate = `{{ .Boilerplate }}

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// PreFlight.DefaultRequeue returns the default requeue configuration for this controller phase.
func (phase *PreFlight) DefaultRequeue() ctrl.Result {
	return controllerutils.DefaultRequeueResult()
}

// PreFlight.Execute executes pre-flight and fail-fast conditions prior to attempting resource creation.
func (phase *PreFlight) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return true, nil
}
`
