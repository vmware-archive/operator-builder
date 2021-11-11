// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Phase{}

// Types scaffolds the phase interfaces and types.
type Phase struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Phase) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "phase.go")

	f.TemplateBody = phaseTemplate

	f.IfExistsAction = machinery.SkipFile

	return nil
}

const phaseTemplate = `{{ .Boilerplate }}

package phases

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

type Handler func(ctx context.Context, r common.ComponentReconciler) (proceed bool, err error)

// Phase defines a phase of the reconciliation process.
type Phase struct {
	Name       string
	definition Handler
}

// DefaultRequeue executes checking for a parent components readiness status.
func (*Phase) DefaultRequeue() ctrl.Result {
	return ctrl.Result{
		Requeue: true,
		// RequeueAfter: 5 * time.Second,
	}
}

// HandlePhaseExit will perform the steps required to exit a phase.
func (p *Phase) HandlePhaseExit(
	ctx context.Context,
	reconciler common.ComponentReconciler,
	phaseIsReady bool,
	phaseError error,
) (ctrl.Result, error) {
	var condition common.PhaseCondition

	var result ctrl.Result

	switch {
	case phaseError != nil:
		if IsOptimisticLockError(phaseError) {
			phaseError = nil
			condition = p.GetSuccessCondition()
		} else {
			condition = p.GetFailCondition(phaseError)
		}

		result = DefaultReconcileResult()
	case !phaseIsReady:
		condition = p.GetPendingCondition()
		result = p.DefaultRequeue()
	default:
		condition = p.GetSuccessCondition()
		result = DefaultReconcileResult()
	}

	// update the status conditions and return any errors
	if updateError := updatePhaseConditions(ctx, reconciler, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !IsOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update status conditions; %v; %w", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	}

	return result, phaseError
}

// GetSuccessCondition defines the success condition for the phase.
func (p *Phase) GetSuccessCondition() common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   p.Name,
		State:   common.PhaseStateComplete,
		Message: "Successfully Completed Phase",
	}
}

// GetPendingCondition defines the pending condition for the phase.
func (p *Phase) GetPendingCondition() common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   p.Name,
		State:   common.PhaseStatePending,
		Message: "Pending Execution of Phase",
	}
}

// GetFailCondition defines the fail condition for the phase.
func (p *Phase) GetFailCondition(err error) common.PhaseCondition {
	return common.PhaseCondition{
		Phase:   p.Name,
		State:   common.PhaseStateFailed,
		Message: "Failed Phase with Error; " + err.Error(),
	}
}

// updatePhaseConditions updates the status.conditions field of the parent custom resource.
func updatePhaseConditions(ctx context.Context, r common.ComponentReconciler, condition *common.PhaseCondition) error {
	r.GetComponent().SetPhaseCondition(condition)

	if err := r.Status().Update(ctx, r.GetComponent()); err != nil {
		return fmt.Errorf("unable to update Phase Condition for %s, %w", r.GetComponent().GetComponentGVK().Kind, err)
	}

	return nil
}
`
