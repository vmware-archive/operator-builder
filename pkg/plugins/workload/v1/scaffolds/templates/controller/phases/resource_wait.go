package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceWait{}

// ResourceWait scaffolds the resource wait phase methods.
type ResourceWait struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourceWait) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource_wait.go")

	f.TemplateBody = resourceWaitTemplate

	return nil
}

const resourceWaitTemplate = `{{ .Boilerplate }}

package phases

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	common "{{ .Repo }}/apis/common"
)

// defaultWaitRequeue defines the default requeue result for this phase.
func defaultWaitRequeue() ctrl.Result {
	return ctrl.Result{Requeue: true}
}

// WaitForResourcePhase.Execute executes waiting for a resource to be ready before continuing.
func (phase *WaitForResourcePhase) Execute(
	resource *ComponentResource,
) (ctrl.Result, bool, error) {
	// TODO: loop through functions instead of repeating logic
	// common wait logic for a resource
	ready, err := commonWait(resource.ComponentReconciler, resource.OriginalResource)

	// return the error if we have any
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return defaultWaitRequeue(), false, nil
	}

	// specific wait logic for a resource
	ready, err = resource.ComponentReconciler.Wait(&resource.OriginalResource)

	// return the error if we have any
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return defaultWaitRequeue(), false, nil
	}

	return ctrl.Result{}, true, nil
}

// commonWait applies all common waiting functions for known resources.
func commonWait(
	r common.ComponentReconciler,
	resource metav1.Object,
) (bool, error) {
	// Namespace
	if resource.GetNamespace() != "" {
		return namespaceIsReady(r, resource)
	}

	return true, nil
}

// namespaceIsReady waits for a namespace object to exist.
func namespaceIsReady(
	r common.ComponentReconciler,
	resource metav1.Object,
) (bool, error) {
	var namespaces v1.NamespaceList

	if err := r.List(r.GetContext(), &namespaces); err != nil {
		return false, err
	}

	// ensure the namespace exists and is not terminating
	for _, namespace := range namespaces.Items {
		if namespace.Name == resource.GetNamespace() {
			if namespace.Status.Phase != v1.NamespaceTerminating {
				return true, nil
			}
		}
	}

	return false, nil
}
`
