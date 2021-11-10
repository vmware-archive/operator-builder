// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceWait{}

// ResourceWait scaffolds the resource wait phase methods.
type ResourceWait struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourceWait) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "resource_wait.go")

	f.TemplateBody = resourceWaitTemplate

	return nil
}

const resourceWaitTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/internal/resources"
)

// WaitForResourcePhase.Execute executes waiting for a resource to be ready before continuing.
func (phase *WaitForResourcePhase) Execute(
	r common.ComponentReconciler,
	resource client.Object,
	resourceCondition *common.ResourceCondition,
) (ctrl.Result, bool, error) {
	// TODO: loop through functions instead of repeating logic
	// common wait logic for a resource
	ready, err := commonWait(r, resource)
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return Requeue(), false, nil
	}

	// specific wait logic for a resource
	meta := resource.(metav1.Object)
	
	ready, err = r.Wait(meta)
	if err != nil {
		return ctrl.Result{}, false, fmt.Errorf("unable to wait for resource %s, %w", meta.GetName(), err)
	}

	// return the result if the object is not ready
	if !ready {
		return Requeue(), false, nil
	}

	return ctrl.Result{}, true, nil
}

// TODO: the following allows all controllers to list all namespaces,
// regardless of whether or not the controller manages namespaces.
//
// This will eventually be moved into a validating webhook so that the user
// will get a message outlining their mistake rather than buried in the
// reconciliation loop, causing pain when having to sift through logs to
// determine a problem.
//
// See:
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/141
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/162

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=list;watch

// commonWait applies all common waiting functions for known resources.
func commonWait(
	r common.ComponentReconciler,
	resource client.Object,
) (bool, error) {
	// Namespace
	if resource.GetNamespace() != "" {
		ready, err := resources.NamespaceForResourceIsReady(r, resource)
		if err != nil {
			return ready, fmt.Errorf("unable to determine if %s namespace is ready, %w", resource.GetNamespace(), err)
		}

		return ready, nil
	}

	return true, nil
}
`
