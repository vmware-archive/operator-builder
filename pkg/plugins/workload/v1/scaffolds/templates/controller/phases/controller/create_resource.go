package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CreateResource{}

// CreateResource scaffolds the create resource phase methods.
type CreateResource struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	IsStandalone bool
}

func (f *CreateResource) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "controller", "create_resource.go")

	f.TemplateBody = createResourceTemplate

	return nil
}

const createResourceTemplate = `{{ .Boilerplate }}

package controller

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	resourcephases "{{ .Repo }}/controllers/phases/resource"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// CreateResources.DefaultRequeue returns the default requeue configuration for this controller phase.
func (phase *CreateResources) DefaultRequeue() ctrl.Result {
	return controllerutils.DefaultRequeueResult()
}

// resourcePhases defines the phases for resource creation and the order in which they run during the reconcile process.
func resourcePhases() []resourcephases.ResourcePhase {
	return []resourcephases.ResourcePhase{
		// wait for other resources before attempting to create
		&resourcephases.Wait{},

		// create the resource in the cluster
		&resourcephases.Persist{},
	}
}

// CreateResources.Execute executes executes sub-phases which are required to create the resources.
func (phase *CreateResources) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	// execute the resource phases against each resource
	for _, resource := range r.GetResources() {
		resourceCommon := resource.ToCommonResource()
		resourceCondition := &common.ResourceCondition{}

		for _, resourcePhase := range resourcePhases() {
			r.GetLogger().V(7).Info(fmt.Sprintf("enter resource phase: %T", resourcePhase))
			_, proceed, err := resourcePhase.Execute(resource, *resourceCondition)

			// set a message, return the error and result on error or when unable to proceed
			if err != nil || !proceed {
				return resourcephases.HandleResourcePhaseExit(
					r,
					*resourceCommon,
					*resourceCondition,
					resourcePhase,
					proceed,
					err,
				)
			}

			// set attributes on the resource condition before updating the status
			resourceCondition.LastResourcePhase = resourcephases.GetResourcePhaseName(resourcePhase)

			r.GetLogger().V(5).Info(fmt.Sprintf("completed resource phase: %T", resourcePhase))
		}
	}

	return true, nil
}
`
