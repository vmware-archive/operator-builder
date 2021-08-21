package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Actions{}

// Actions scaffolds controller utilities common to all controllers.
type Actions struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Actions) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "utils", "actions.go")

	f.TemplateBody = actionsTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const actionsTemplate = `{{ .Boilerplate }}

package utils

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"{{ .Repo }}/apis/common"
)

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
				OwnerType:    r.GetComponent().(runtime.Object),
			},
			ResourcePredicates(r),
		); err != nil {
			return err
		}

		r.SetWatch(resource)
	}

	return nil
}
`
