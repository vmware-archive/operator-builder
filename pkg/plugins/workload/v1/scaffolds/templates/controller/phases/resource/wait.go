package resource

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Wait{}

// Wait scaffolds the resource wait phase methods.
type Wait struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Wait) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource", "wait.go")

	f.TemplateBody = waitTemplate

	return nil
}

const waitTemplate = `{{ .Boilerplate }}

package resource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/pkg/resources"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// Wait.Execute executes waiting for a resource to be ready before continuing.
func (phase *Wait) Execute(
	resource common.ComponentResource,
	resourceCondition common.ResourceCondition,
) (ctrl.Result, bool, error) {
	// TODO: loop through functions instead of repeating logic
	// common wait logic for a resource
	ready, err := commonWait(resource.GetReconciler(), resource)
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return controllerutils.DefaultRequeueResult(), false, nil
	}

	// specific wait logic for a resource
	meta := resource.GetObject().(metav1.Object)
	ready, err = resource.GetReconciler().Wait(&meta)
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return controllerutils.DefaultRequeueResult(), false, nil
	}

	return ctrl.Result{}, true, nil
}

// commonWait applies all common waiting functions for known resources.
func commonWait(
	r common.ComponentReconciler,
	resource common.ComponentResource,
) (bool, error) {
	// Namespace
	if resource.GetObject().GetNamespace() != "" {
		return resources.NamespaceForResourceIsReady(resource)
	}

	return true, nil
}
`
