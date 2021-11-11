// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Registry{}

// Types scaffolds the phase interfaces and types.
type Registry struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Registry) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "registry.go")

	f.TemplateBody = registryTemplate

	return nil
}

const registryTemplate = `{{ .Boilerplate }}

package phases

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"{{ .Repo }}/apis/common"
)

type Event int32

const (
	CreateEvent Event = iota
	UpdateEvent
	DeleteEvent
)

type Registry struct {
	createPhases []*Phase
	updatePhases []*Phase
	deletePhases []*Phase
}

func (registry *Registry) Register(name string, definition Handler, event Event) {
	phase := &Phase{
		Name:       name,
		definition: definition,
	}

	switch event {
	case CreateEvent:
		registry.createPhases = append(registry.createPhases, phase)
	case UpdateEvent:
		registry.updatePhases = append(registry.updatePhases, phase)
	case DeleteEvent:
		registry.deletePhases = append(registry.deletePhases, phase)
	}
}

func (registry *Registry) Execute(ctx context.Context, r common.ComponentReconciler, event Event) (reconcile.Result, error) {
	phases := registry.getPhases(event)
	for _, phase := range phases {
		r.GetLogger().V(7).Info(
			"enter phase",
			"phase", phase.Name,
		)

		proceed, err := phase.definition(ctx, r)
		result, err := phase.HandlePhaseExit(ctx, r, proceed, err)

		if err != nil || !proceed {
			r.GetLogger().V(2).Info(
				"not ready; requeuing",
				"phase", phase.Name,
			)

			// return only if we have an error or are told not to proceed
			if err != nil {
				return result, fmt.Errorf("unable to complete %s phase for %s, %w", phase.Name, r.GetComponent().GetComponentGVK().Kind, err)
			}

			if !proceed {
				return result, nil
			}
		}

		r.GetLogger().V(5).Info(
			"completed phase",
			"phase", phase.Name,
		)
	}

	return DefaultReconcileResult(), nil
}

func (registry *Registry) getPhases(event Event) []*Phase {
	switch event {
	case CreateEvent:
		return registry.createPhases
	case UpdateEvent:
		return registry.updatePhases
	case DeleteEvent:
		return registry.deletePhases
	}

	return nil
}

func (registry *Registry) HandleDelete(ctx context.Context, r common.ComponentReconciler) (ctrl.Result, error) {
	myFinalizerName := fmt.Sprintf("%s/Finalizer", r.GetComponent().GetComponentGVK().Group)

	if r.GetComponent().GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	// The object is being deleted
	if containsString(r.GetComponent().GetFinalizers(), myFinalizerName) {
		// our finalizer is present, so lets handle any external dependency
		result, err := registry.Execute(ctx, r, DeleteEvent)
		if err != nil || !result.IsZero() {
			return result, err
		}

		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(r.GetComponent(), myFinalizerName)

		if err := r.Update(ctx, r.GetComponent()); err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to remove finalizer from %s, %w", r.GetComponent().GetComponentGVK().Kind, err)
		}
	}

	// Stop reconciliation as the item is being deleted
	return ctrl.Result{}, nil
}
`
