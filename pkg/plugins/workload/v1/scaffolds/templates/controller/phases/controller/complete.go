package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Complete{}

// Complete scaffolds the complete phase methods.
type Complete struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Complete) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "controller", "complete.go")

	f.TemplateBody = completeTemplate

	return nil
}

const completeTemplate = `{{ .Boilerplate }}

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// Complete.DefaultRequeue returns the default requeue configuration for this controller phase.
func (phase *Complete) DefaultRequeue() ctrl.Result {
	return controllerutils.DefaultRequeueResult()
}

// Complete.Execute executes the completion of a reconciliation loop.
func (phase *Complete) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	r.GetComponent().SetReadyStatus(true)
	r.GetLogger().V(0).Info("successfully reconciled")

	return true, nil
}
`
