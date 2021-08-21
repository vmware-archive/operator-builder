package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Dependency{}

// Dependency scaffolds the dependency phase methods.
type Dependency struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Dependency) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "controller", "dependency.go")

	f.TemplateBody = dependencyTemplate

	return nil
}

const dependencyTemplate = `{{ .Boilerplate }}

package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/pkg/helpers"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// Dependency.DefaultRequeue returns the default requeue configuration for this controller phase.
func (phase *Dependency) DefaultRequeue() ctrl.Result {
	return controllerutils.DefaultRequeueResult()
}

// Dependency.Execute executes a dependency check prior to attempting to create resources.
func (phase *Dependency) Execute(r common.ComponentReconciler) (proceedToNextPhase bool, err error) {
	// dependencies
	component := r.GetComponent()

	if !collectionConfigIsReady(r) {
		return false, nil
	}

	// TODO: set DependenciesSatisfied field (see next TODO below)
	if !component.GetDependencyStatus() {
		satisfied, err := dependenciesSatisfied(r)
		if err != nil || !satisfied {
			return false, err
		}

		// dependencies satisfied; set and update status and continue
		// TODO: needs implemented
	}

	return true, nil
}

// dependenciesSatisfied will return whether or not all dependencies are satisfied for a component.
func dependenciesSatisfied(r common.ComponentReconciler) (bool, error) {
	for _, dep := range r.GetComponent().GetDependencies() {
		satisfied, err := dependencySatisfied(r, dep)
		if err != nil || !satisfied {
			return false, err
		}
	}

	return true, nil
}

// dependencySatisfied will return whether or not an individual dependency is satisfied.
func dependencySatisfied(r common.ComponentReconciler, dependency common.Component) (bool, error) {
	// get the dependencies by kind that already exist in cluster
	dependencyList := &unstructured.UnstructuredList{}

	dependencyList.SetGroupVersionKind(dependency.GetComponentGVK())

	if err := r.List(r.GetContext(), dependencyList, &client.ListOptions{}); err != nil {
		return false, err
	}

	// expect only one item returned, otherwise dependencies are considered unsatisfied
	if len(dependencyList.Items) != 1 {
		return false, nil
	}

	// get the status.created field on the object and return the status and any errors found
	status, found, err := unstructured.NestedBool(dependencyList.Items[0].Object, "status", "created")
	if err != nil || !found {
		return false, err
	}

	return status, nil
}

// collectionConfigIsReady determines if a component's collection is ready.
func collectionConfigIsReady(r common.ComponentReconciler) bool {
	// get a list of configurations from the cluster
	collectionConfigs, err := helpers.GetCollectionConfigs(r)
	if err != nil {
		r.GetLogger().V(0).Info("unable to find resource of kind: [" + helpers.CollectionAPIKind + "]")

		return false
	}

	// configuration is not ready if we do not have exactly one configuration
	if len(collectionConfigs.Items) != 1 {
		r.GetLogger().V(0).Info(
			fmt.Sprintf("expected only 1 resource of kind: [%s]; found %v", helpers.CollectionAPIKind, len(collectionConfigs.Items)),
		)

		return false
	}

	return true
}
`
