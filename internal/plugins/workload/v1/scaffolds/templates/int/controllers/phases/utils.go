// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

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
	f.Path = filepath.Join("internal", "controllers", "phases", "utils.go")

	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"{{ .Repo }}/apis/common"
)

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

// Requeue will return the default result to requeue a reconciler request when needed.
func Requeue() ctrl.Result {
	return ctrl.Result{Requeue: true}
}

// IsOptimisticLockError checks to see if the error is a locking error.
func IsOptimisticLockError(err error) bool {
	return strings.Contains(err.Error(), optimisticLockErrorMsg)
}

// DefaultReconcileResult will return the default reconcile result when requeuing is not needed.
func DefaultReconcileResult() ctrl.Result {
	return ctrl.Result{}
}

func RegisterDeleteHooks(r common.ComponentReconciler) error {
	myFinalizerName := fmt.Sprintf("%s/Finalizer", r.GetComponent().GetComponentGVK().Group)

	if r.GetComponent().GetDeletionTimestamp().IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(r.GetComponent().GetFinalizers(), myFinalizerName) {
			controllerutil.AddFinalizer(r.GetComponent(), myFinalizerName)

			if err := r.Update(r.GetContext(), r.GetComponent()); err != nil {
				return fmt.Errorf("unable to register delete hook on %s, %w", r.GetComponent().GetComponentGVK().Kind, err)
			}
		}
	}

	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}

	return false
}
`
