// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package utils

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Utils{}

// Utils scaffolds controller utilities common to all controllers.
type Utils struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	IsStandalone bool
}

func (f *Utils) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "utils", "utils.go")

	f.TemplateBody = controllerUtilsTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const controllerUtilsTemplate = `{{ .Boilerplate }}

package utils

import (
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"{{ .Repo }}/apis/common"

	"github.com/nukleros/operator-builder-tools/pkg/resources"
)

const (
	FieldManager = "reconciler"
)

// GetDesiredObject returns the desired object from a list stored in memory.
func GetDesiredObject(compared client.Object, r common.ComponentReconciler) (client.Object, error) {
	var desired client.Object

	allObjects, err := r.GetResources()
	if err != nil {
		return nil, fmt.Errorf("unable to get resources, %w", err)
	}

	for _, resource := range allObjects {
		if resources.EqualGVK(compared, resource) && resources.EqualNamespaceName(compared, resource) {
			return resource, nil
		}
	}

	return desired, nil
}

// needsReconciliation performs some simple checks and returns whether or not a
// resource needs to be updated.
func needsReconciliation(r common.ComponentReconciler, existing, requested client.Object) bool {
	// skip if the resources versions are the same
	if existing.GetResourceVersion() == requested.GetResourceVersion() {
		return false
	}

	// get the desired object from the reconciler and ensure that we both
	// found that desired object and that the desired object fields are equal
	// to the existing object fields
	desired, err := GetDesiredObject(requested, r)
	if err != nil {
		r.GetLogger().V(0).Error(
			err, "unable to get object in memory",
			"kind", requested.GetObjectKind().GroupVersionKind().Kind,
			"name", requested.GetName(),
			"namespace", requested.GetNamespace(),
		)

		return false
	}

	if desired == nil {
		return true
	}

	equal, err := resources.AreEqual(desired, requested)
	if err != nil {
		r.GetLogger().V(0).Error(
			err, "unable to determine equality for reconciliation",
			"kind", desired.GetObjectKind().GroupVersionKind().Kind,
			"name", desired.GetName(),
			"namespace", desired.GetNamespace(),
		)

		return true
	}

	return !equal
}

// ResourcePredicates returns the filters which are used to filter out the common reconcile events
// prior to reconciling the child resource of a component.
func ResourcePredicates(r common.ComponentReconciler) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return needsReconciliation(
				r,
				e.ObjectOld,
				e.ObjectNew,
			)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			// do not run reconciliation on unknown events
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			// do not run reconciliation again when we just created the child resource
			return false
		},
	}
}

// ComponentPredicates returns the filters which are used to filter out the common reconcile events
// prior to reconciling an object for a component.
func ComponentPredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return !e.DeleteStateUnknown
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// Watch watches a resource.
func Watch(
	r common.ComponentReconciler,
	resource client.Object,
) error {
	// check if the resource is already being watched
	var watched bool

	if len(r.GetWatches()) > 0 {
		for _, watcher := range r.GetWatches() {
			if reflect.DeepEqual(
				resource.GetObjectKind().GroupVersionKind(),
				watcher.GetObjectKind().GroupVersionKind(),
			) {
				watched = true
			}
		}
	}

	// watch the resource if it current is not being watched
	if !watched {
		if err := r.GetController().Watch(
			&source.Kind{Type: resource},
			&handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    r.GetComponent(),
			},
			ResourcePredicates(r),
		); err != nil {
			return fmt.Errorf("unable to watch resource, %w", err)
		}

		r.SetWatch(resource)
	}

	return nil
}
`
