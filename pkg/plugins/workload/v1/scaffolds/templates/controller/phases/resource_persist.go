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

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// PersistResourcePhase.Execute executes persisting resources to the Kubernetes database.
<<<<<<< HEAD
func (phase *PersistResourcePhase) Execute(
	resource common.ComponentResource,
	resourceCondition common.ResourceCondition,
) (ctrl.Result, bool, error) {
	// persist the resource
	if err := persistResource(
		resource,
		resourceCondition,
		phase,
	); err != nil {
		return ctrl.Result{}, false, err
=======
func (phase *PersistResourcePhase) Execute(resource *ComponentResource) (ctrl.Result, bool, error) {
	// if we are skipping resource creation, return immediately
	if resource.Skip {
		return ctrl.Result{}, true, nil
	}

	// if we are replacing resources, use the replaced resources, else use the original resources
	if len(resource.ReplacedResources) > 0 {
		for _, replacedResource := range resource.ReplacedResources {
			if err := persistResource(resource.ComponentReconciler, replacedResource); err != nil {
				return ctrl.Result{}, false, err
			}
		}
	} else {
		if err := persistResource(resource.ComponentReconciler, resource.OriginalResource); err != nil {
			return ctrl.Result{}, false, err
		}
>>>>>>> upstream/main
	}

	return ctrl.Result{}, true, nil
}

// persistResource persists a single resource to the Kubernetes database.
func persistResource(
	resource common.ComponentResource,
	condition common.ResourceCondition,
	phase *PersistResourcePhase,
) error {
	// persist resource
<<<<<<< HEAD
	r := resource.GetReconciler()
	if err := r.CreateOrUpdate(resource.GetObject()); err != nil {
		if IsOptimisticLockError(err) {
			return nil
		} else {
			r.GetLogger().V(0).Info(err.Error())
=======
	if err := r.CreateOrUpdate(resource); err != nil {
		if isOptimisticLockError(err) {
			return nil
		} else {
			r.GetLogger().V(0).Info("failed persisting object of kind: " + objectKind + " with name: " + objectName)
>>>>>>> upstream/main

			return err
		}
	}

	// set attributes related to the persistence of this child resource
	condition.LastResourcePhase = getResourcePhaseName(phase)
	condition.LastModified = time.Now().UTC().String()
	condition.Message = "resource created successfully"
	condition.Created = true

	// update the condition to notify that we have created a child resource
	return updateResourceConditions(r, *resource.ToCommonResource(), &condition)
}
`
