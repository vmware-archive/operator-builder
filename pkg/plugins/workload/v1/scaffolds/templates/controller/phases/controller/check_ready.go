package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CheckReady{}

// CheckReady scaffolds the check ready phase methods.
type CheckReady struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *CheckReady) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "controller", "check_ready.go")

	f.TemplateBody = checkReadyTemplate

	return nil
}

const checkReadyTemplate = `{{ .Boilerplate }}

package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/pkg/resources"
)

// CheckReady.DefaultRequeue returns the default requeue configuration for this controller phase.
func (phase *CheckReady) DefaultRequeue() ctrl.Result {
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 5 * time.Second,
	}
}

// CheckReady.Execute executes checking for a parent components readiness status.
func (phase *CheckReady) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	// check to see if known types are ready
	knownReady, err := resources.AreReady(r.GetResources()...)
	if err != nil {
		return false, err
	}

	// check to see if the custom methods return ready
	customReady, err := r.CheckReady()
	if err != nil {
		return false, err
	}

	return (knownReady && customReady), nil
}
`
