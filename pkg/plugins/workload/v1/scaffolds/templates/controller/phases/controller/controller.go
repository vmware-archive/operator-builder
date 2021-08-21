package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Controller{}

// Controller scaffolds common phase operations.
type Controller struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Controller) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "controller", "controller.go")

	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package controller

import (
	"fmt"
	"reflect"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	controllerutils "{{ .Repo }}/controllers/utils"
)

// CreatePhases defines the phases for create and the order in which they run during the reconcile process.
func CreatePhases() []Phase {
	return []Phase{
		&Dependency{},
		&PreFlight{},
		&CreateResources{},
		&CheckReady{},
		&Complete{},
	}
}

// UpdatePhases defines the phases for update and the order in which they run during the reconcile process.
func UpdatePhases() []Phase {
	return []Phase{
		&Dependency{},
		&PreFlight{},
		&CreateResources{},
		&CheckReady{},
		&Complete{},
	}
}

// GetPhases returns which phases to run given the component.
func GetPhases(component common.Component) []Phase {
	var phases []Phase
	if !component.GetReadyStatus() {
		phases = CreatePhases()
	} else {
		phases = UpdatePhases()
	}

	return phases
}

// GetSuccessCondition defines the success condition for the phase.
func GetSuccessCondition(phase Phase) common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   getPhaseName(phase),
		State:   common.PhaseStateComplete,
		Message: "Successfully Completed Phase",
	}
}

// GetPendingCondition defines the pending condition for the phase.
func GetPendingCondition(phase Phase) common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   getPhaseName(phase),
		State:   common.PhaseStatePending,
		Message: "Pending Execution of Phase",
	}
}

// GetFailCondition defines the fail condition for the phase.
func GetFailCondition(phase Phase, err error) common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   getPhaseName(phase),
		State:   common.PhaseStateFailed,
		Message: "Failed Phase with Error; " + err.Error(),
	}
}

func getPhaseName(phase Phase) string {
	objectElements := strings.Split(fmt.Sprintf("%s", reflect.TypeOf(phase)), ".")

	return objectElements[len(objectElements)-1]
}

// updatePhaseConditions updates the status.conditions field of the parent custom resource.
func updatePhaseConditions(
	r common.ComponentReconciler,
	condition *common.PhaseCondition,
) error {
	r.GetComponent().SetPhaseCondition(*condition)

	return r.UpdateStatus()
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
		if controllerutils.IsOptimisticLockError(phaseError) {
			phaseError = nil
			condition = GetSuccessCondition(phase)
		} else {
			condition = GetFailCondition(phase, phaseError)
		}
		result = controllerutils.DefaultReconcileResult()
	case !phaseIsReady:
		condition = GetPendingCondition(phase)
		result = phase.DefaultRequeue()
	default:
		condition = GetSuccessCondition(phase)
		result = controllerutils.DefaultReconcileResult()
	}

	// update the status conditions and return any errors
	if updateError := updatePhaseConditions(reconciler, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !controllerutils.IsOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update status conditions; %v; %v", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	}

	return result, phaseError
}
`
