package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourcePersist{}

// ResourcePersist scaffolds the resource persist phase methods.
type ResourcePersist struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourcePersist) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource_persist.go")

	f.TemplateBody = resourcePersistTemplate

	return nil
}

const resourcePersistTemplate = `{{ .Boilerplate }}

package phases

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// PersistResourcePhase.Execute executes persisting resources to the Kubernetes database.
func (phase *PersistResourcePhase) Execute(resource *ComponentResource) (ctrl.Result, bool, error) {
	// if we are skipping resource creation, return immediately
	if resource.Skip {
		return ctrl.Result{}, true, nil
	}

	// if we are replacing resources, use the replaced resources, else use the original resources
	var resources []metav1.Object
	if len(resource.ReplacedResources) > 0 {
		resources = resource.ReplacedResources
	} else {
		resources = []metav1.Object{resource.OriginalResource}
	}

	// loop through the resources and persist as necessary
	for _, resourceObject := range resources {
		if err := persistResource(
			resource.ComponentReconciler,
			resourceObject,
			resource.ResourceCondition,
			phase,
		); err != nil {
			return ctrl.Result{}, false, err
		}
	}

	return ctrl.Result{}, true, nil
}

// persistResource persists a single resource to the Kubernetes database.
func persistResource(
	r common.ComponentReconciler,
	resource metav1.Object,
	condition common.ResourceCondition,
	phase *PersistResourcePhase,
) error {
	// persist resource
	if err := r.CreateOrUpdate(resource); err != nil {
		if isOptimisticLockError(err) {
			return nil
		} else {
			return err
		}
	}

	// set attributes related to the persistence of this child resource
	condition.LastResourcePhase = getResourcePhaseName(phase)
	condition.LastModified = time.Now().UTC().String()
	condition.Message = "resource created successfully"
	condition.Created = true

	// update the condition to notify that we have created a child resource
	return updateResourceConditions(r, &condition)
}
`
