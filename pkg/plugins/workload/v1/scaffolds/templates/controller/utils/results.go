package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Results{}

// Results scaffolds common phase operations.
type Results struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Results) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "utils", "results.go")

	f.TemplateBody = resultsTemplate

	return nil
}

const resultsTemplate = `{{ .Boilerplate }}

package utils

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

// DefaultRequeueResult will return the default result to requeue a reconciler request when needed.
func DefaultRequeueResult() ctrl.Result {
	return ctrl.Result{Requeue: true}
}

// DefaultReconcileResult will return the default reconcile result when requeuing is not needed.
func DefaultReconcileResult() ctrl.Result {
	return ctrl.Result{}
}
`
