package common

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Components{}

// Components scaffolds the interfaces between workloads
type Components struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
}

func (f *Components) SetTemplateDefaults() error {

	f.Path = filepath.Join("apis", "common", "components.go")

	f.TemplateBody = commonTemplate

	return nil
}

var commonTemplate = `
// +build !ignore_autogenerated

{{ .Boilerplate }}
package common

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//runtime "k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	//"k8s.io/apimachinery/pkg/types"
	//"sigs.k8s.io/controller-runtime/pkg/client"
)

type Component interface {
	GetReadyStatus() bool
	//SetReadyStatus(bool)
	//GetDependencyStatus() bool
	//SetDependencyStatus(bool)
	GetStatusConditions() []Condition
	SetStatusConditions(Condition)
	//GetDependencies() []Component
	//GetComponentGVK() schema.GroupVersionKind
}

type ComponentReconciler interface {
	//List(context.Context, runtime.Object, ...client.ListOption) error
	//Get(context.Context, types.NamespacedName, runtime.Object) error
	//GetClient() client.Client
	//GetScheme() *runtime.Scheme
	GetContext() context.Context
	GetComponent() Component
	GetLogger() logr.Logger
	GetResources(Component) ([]metav1.Object, error)
	SetRefAndCreateIfNotPresent(metav1.Object) error
	UpdateStatus(context.Context, Component) error
	//CheckReady() (bool, error)
	//Mutate(*metav1.Object) ([]metav1.Object, bool, error)
	//Wait(*metav1.Object) (bool, error)
}
`