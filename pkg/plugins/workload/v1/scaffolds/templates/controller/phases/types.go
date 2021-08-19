package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Types{}

// Types scaffolds the phase interfaces and types.
type Types struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Types) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "types.go")

	f.TemplateBody = typesTemplate

	return nil
}

const typesTemplate = `{{ .Boilerplate }}

package phases

import (
<<<<<<< HEAD
	"fmt"
	"reflect"
	"strings"

=======
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
>>>>>>> upstream/main
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// Phase defines a phase of the reconciliation process.
type Phase interface {
	Execute(common.ComponentReconciler) (bool, error)
<<<<<<< HEAD
	DefaultRequeue() ctrl.Result
=======
}

// PhaseHandler defines an object which can handle the outcome of a phase execution.
type PhaseHandler interface {
	GetSuccessCondition() common.Condition
	GetPendingCondition() common.Condition
	GetFailCondition()    common.Condition
>>>>>>> upstream/main
}

// ResourcePhase defines the specific phase of reconcilication associated with creating resources.
type ResourcePhase interface {
	Execute(common.ComponentResource, common.ResourceCondition) (ctrl.Result, bool, error)
}

<<<<<<< HEAD
// Below are the phase types which satisfy the Phase interface.
=======
// ComponentResource defines a resource which is created by the parent Component custom resource.
type ComponentResource struct {
	ComponentReconciler common.ComponentReconciler
	OriginalResource  metav1.Object
	ReplacedResources []metav1.Object
	Skip              bool
}

// DependencyPhase defines an object specific to the depenency phase of reconciliation.
>>>>>>> upstream/main
type DependencyPhase struct{}
type PreFlightPhase struct{}
type CreateResourcesPhase struct{}
type CheckReadyPhase struct{}
type CompletePhase struct{}

// Below are the phase types which satisfy the ResourcePhase interface.
type PersistResourcePhase struct{}
type WaitForResourcePhase struct{}

<<<<<<< HEAD
// GetSuccessCondition defines the success condition for the phase.
func GetSuccessCondition(phase Phase) common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   getPhaseName(phase),
		State:   common.PhaseStateComplete,
		Message: "Successfully Completed Phase",
	}
=======
// CreateResourcesPhase defines an object specific to the create resources phase of reconciliation.
type CreateResourcesPhase struct {
	Resources         []metav1.Object
	ReplacedResources []metav1.Object
>>>>>>> upstream/main
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

func getResourcePhaseName(resourcePhase ResourcePhase) string {
	objectElements := strings.Split(fmt.Sprintf("%s", reflect.TypeOf(resourcePhase)), ".")

	return objectElements[len(objectElements)-1]
}
`
