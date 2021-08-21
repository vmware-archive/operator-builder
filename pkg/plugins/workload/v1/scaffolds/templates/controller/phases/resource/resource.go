package resource

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Resource{}

// Resource scaffolds common phase operations.
type Resource struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Resource) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource", "resource.go")

	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package resource

import (
	"fmt"
	"reflect"
	"strings"

	"{{ .Repo }}/apis/common"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// updateResourceConditions updates the status.resourceConditions field of the parent custom resource.
func updateResourceConditions(
	r common.ComponentReconciler,
	resource common.Resource,
	condition *common.ResourceCondition,
) error {
	resource.ResourceCondition = *condition
	r.GetComponent().SetResource(resource)

	return r.UpdateStatus()
}

// GetResourcePhaseName returns the proper name of the resource phase.
func GetResourcePhaseName(resourcePhase ResourcePhase) string {
	objectElements := strings.Split(fmt.Sprintf("%s", reflect.TypeOf(resourcePhase)), ".")

	return objectElements[len(objectElements)-1]
}

// HandleResourcePhaseExit will perform the steps required to exit a phase.
func HandleResourcePhaseExit(
	reconciler common.ComponentReconciler,
	resource common.Resource,
	condition common.ResourceCondition,
	phase ResourcePhase,
	phaseIsReady bool,
	phaseError error,
) (bool, error) {

	switch {
	case phaseError != nil:
		if controllerutils.IsOptimisticLockError(phaseError) {
			phaseError = nil
		}
	case !phaseIsReady:
		condition.Message = fmt.Sprintf("unable to proceed with resource creation; phase %v is not ready", GetResourcePhaseName(phase))
	}

	// update the status conditions and return any errors
	if updateError := updateResourceConditions(reconciler, resource, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !controllerutils.IsOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update resource conditions; %v; %v", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	} else {
		condition.Message = "resource creation successful"
	}

	return (phaseError == nil && phaseIsReady), phaseError
}
`
