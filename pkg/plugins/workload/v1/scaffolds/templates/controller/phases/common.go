package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Common{}

// Common scaffolds common phase operations.
type Common struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Common) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "common.go")

	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"
	"time"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

// Requeue will return the default result to requeue a reconciler request when needed.
func Requeue() ctrl.Result {
	return ctrl.Result{Requeue: true}
}

// DefaultReconcileResult will return the default reconcile result when requeuing is not needed.
func DefaultReconcileResult() ctrl.Result {
	return ctrl.Result{}
}

// phaseConditionExists will return whether or not a specific condition already exists on the object.
func phaseConditionExists(
	currentConditions []common.PhaseCondition,
	condition *common.PhaseCondition,
) bool {
	for _, currentCondition := range currentConditions {
		if reflect.DeepEqual(currentCondition, *condition) {
			return true
		}
	}

	return false
}

// resourceConditionExists will return whether or not a specific resource condition already exists on the object.
func resourceConditionExists(
	currentConditions []common.ResourceCondition,
	condition *common.ResourceCondition,
) bool {
	for _, currentCondition := range currentConditions {
		if reflect.DeepEqual(currentCondition, *condition) {
			return true
		}
	}

	return false
}

// updatePhaseConditions updates the status.conditions field of the parent custom resource.
func updatePhaseConditions(
	r common.ComponentReconciler,
	condition *common.PhaseCondition,
) error {
	component := r.GetComponent()

	if !phaseConditionExists(component.GetPhaseConditions(), condition) {
		if condition.LastModified == "" {
			condition.LastModified = time.Now().UTC().String()
		}
		component.SetPhaseCondition(*condition)

		return r.UpdateStatus()
	}

	return nil
}

// updateResourceConditions updates the status.resourceConditions field of the parent custom resource.
func updateResourceConditions(
	r common.ComponentReconciler,
	condition *common.ResourceCondition,
) error {
	component := r.GetComponent()

	if !resourceConditionExists(component.GetResourceConditions(), condition) {
		if condition.LastModified == "" {
			condition.LastModified = time.Now().UTC().String()
		}
		component.SetResourceCondition(*condition)

		return r.UpdateStatus()
	}

	return nil
}

// HandlePhaseExit will perform the steps required to exit a phase.
func HandlePhaseExit(
	reconciler common.ComponentReconciler,
	phase Phase,
	phaseIsReady bool,
	phaseError error,
) (ctrl.Result, error) {
	var condition common.PhaseCondition

	var result ctrl.Result

	switch {
	case phaseError != nil:
		if isOptimisticLockError(phaseError) {
			phaseError = nil
			condition = GetSuccessCondition(phase)
		} else {
			condition = GetFailCondition(phase, phaseError)
		}
		result = DefaultReconcileResult()
	case !phaseIsReady:
		condition = GetPendingCondition(phase)
		result = Requeue()
	default:
		condition = GetSuccessCondition(phase)
		result = DefaultReconcileResult()
	}

	// update the status conditions and return any errors
	if updateError := updatePhaseConditions(reconciler, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !isOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update status conditions; %v; %v", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	}

	return result, phaseError
}

// handleResourcePhaseExit will perform the steps required to exit a phase.
func handleResourcePhaseExit(
	reconciler common.ComponentReconciler,
	resource client.Object,
	condition common.ResourceCondition,
	phaseIsReady bool,
	phaseError error,
) (bool, error) {

	switch {
	case phaseError != nil:
		if isOptimisticLockError(phaseError) {
			phaseError = nil
			condition.Message = "resource creation successful"
		}
	case !phaseIsReady:
		condition.Message = "unable to proceed with resource creation"
	default:
		condition.Message = "resource creation successful"
	}

	// update the status conditions and return any errors
	if updateError := updateResourceConditions(reconciler, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !isOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update resource conditions; %v; %v", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	}

	return (phaseError == nil && phaseIsReady), phaseError
}

// isOptimisticLockError checks to see if the error is a locking error.
func isOptimisticLockError(err error) bool {
	return strings.Contains(err.Error(), optimisticLockErrorMsg)
}

// setResources will set the resources against a CreateResourcePhase.
func setResources(
	parent *CreateResourcesPhase,
	resources []metav1.Object,
) {
	parent.Resources = resources
}

// getResources will get the resources from a CreateResourcePhase.
func getResources(
	parent *CreateResourcesPhase,
) []metav1.Object {
	return parent.Resources
}
`
