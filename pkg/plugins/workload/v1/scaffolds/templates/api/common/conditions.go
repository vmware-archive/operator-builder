package common

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Conditions{}

// Conditions scaffolds the conditions for all workloads.
type Conditions struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
}

func (f *Conditions) SetTemplateDefaults() error {
	f.Path = filepath.Join("apis", "common", "conditions.go")

	f.TemplateBody = conditionsTemplate

	return nil
}

const conditionsTemplate = `{{ .Boilerplate }}

package common

// PhaseState defines the current state of the phase.
// +kubebuilder:validation:Enum=Complete;Reconciling;Failed;Pending
type PhaseState string

const (
	PhaseStatePending     PhaseState = "Pending"
	PhaseStateReconciling PhaseState = "Reconciling"
	PhaseStateFailed      PhaseState = "Failed"
	PhaseStateComplete    PhaseState = "Complete"
)

// PhaseCondition describes an event that has occurred during a phase
// of the controller reconciliation loop.
type PhaseCondition struct {
	State PhaseState ` + "`" + `json:"state"` + "`" + `

	// Phase defines the phase in which the condition was set.
	Phase string ` + "`" + `json:"phase"` + "`" + `

	// Message defines a helpful message from the phase.
	Message string ` + "`" + `json:"message"` + "`" + `

	// LastModified defines the time in which this component was updated.
	LastModified string ` + "`" + `json:"lastModified"` + "`" + `
}

// ResourceCondition describes that condition of a Kubernetes resource managed by the parent object.
type ResourceCondition struct {
	// Group defines the API Group of the resource that this status applies to.
	Group string ` + "`" + `json:"group"` + "`" + `

	// Version defines the API Version of the resource that this status applies to.
	Version string ` + "`" + `json:"version"` + "`" + `

	// Kind defines the kind of resource that this status applies to.
	Kind string ` + "`" + `json:"kind"` + "`" + `

	// Name defines the name of the resource from the metadata.name field.
	Name string ` + "`" + `json:"name"` + "`" + `

	// Namespace defines the namespace in which this resource exists in.
	Namespace string ` + "`" + `json:"namespace"` + "`" + `

	// Created defined whether this object has been successfully created or not.
	Created bool ` + "`" + `json:"created"` + "`" + `

	// LastResourcePhase defines the last successfully completed resource phase.
	LastResourcePhase string ` + "`" + `json:"lastResourcePhase"` + "`" + `

	// LastModified defines the time in which this resource was updated.
	LastModified string ` + "`" + `json:"lastModified"` + "`" + `

	// Message defines a helpful message from the resource phase.
	Message string ` + "`" + `json:"message"` + "`" + `
}
`
