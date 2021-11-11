// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourcePhases{}

// Types scaffolds the phase interfaces and types.
type ResourcePhases struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourcePhases) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "resource_phase.go")

	f.TemplateBody = resourcePhaseTemplate

	return nil
}

const resourcePhaseTemplate = `{{ .Boilerplate }}

package phases

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
)

// ResourcePhase defines the specific phase of reconcilication associated with creating resources.
type ResourcePhase interface {
	Execute(context.Context, common.ComponentReconciler, client.Object, *common.ResourceCondition) (ctrl.Result, bool, error)
}

// Below are the phase types which satisfy the ResourcePhase interface.
type (
	PersistResourcePhase struct{}
	WaitForResourcePhase struct{}
)

// updateResourceConditions updates the status.resourceConditions field of the parent custom resource.
func updateResourceConditions(
	ctx context.Context,
	r common.ComponentReconciler,
	resource *common.Resource,
	condition *common.ResourceCondition,
) error {
	resource.ResourceCondition = *condition
	r.GetComponent().SetResource(resource)

	if err := r.Status().Update(ctx, r.GetComponent()); err != nil {
		return fmt.Errorf("unable to update Resource Condition for %s, %w", r.GetComponent().GetComponentGVK().Kind, err)
	}

	return nil
}


// handleResourcePhaseExit will perform the steps required to exit a phase.
func handleResourcePhaseExit(
	ctx context.Context,
	reconciler common.ComponentReconciler,
	resource *common.Resource,
	condition *common.ResourceCondition,
	phase ResourcePhase,
	phaseIsReady bool,
	phaseError error,
) (bool, error) {
	switch {
	case phaseError != nil:
		if IsOptimisticLockError(phaseError) {
			phaseError = nil
		}
	case !phaseIsReady:
		condition.Message = fmt.Sprintf("unable to proceed with resource creation; phase %v is not ready", getResourcePhaseName(phase))
	default:
		condition.Message = "resource creation successful"
	}

	// update the status conditions and return any errors
	if updateError := updateResourceConditions(ctx, reconciler, resource, condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !IsOptimisticLockError(updateError) {
			if phaseError != nil {
				return false, fmt.Errorf("failed to update resource conditions; %v; %w", updateError, phaseError)
			}

			return false, updateError
		}
	}

	return (phaseError == nil && phaseIsReady), phaseError
}

func getResourcePhaseName(resourcePhase ResourcePhase) string {
	objectElements := strings.Split(reflect.TypeOf(resourcePhase).String(), ".")

	return objectElements[len(objectElements)-1]
}
`
