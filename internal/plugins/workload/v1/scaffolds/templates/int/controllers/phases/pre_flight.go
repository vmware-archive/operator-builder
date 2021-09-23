// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PreFlight{}

// PreFlight scaffolds the pre-flight phase methods.
type PreFlight struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *PreFlight) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "pre_flight.go")

	f.TemplateBody = preFlightTemplate

	return nil
}

const preFlightTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// PreFlightPhase.DefaultRequeue defines the default requeue result for this
// phase.
func (phase *PreFlightPhase) DefaultRequeue() ctrl.Result {
	return Requeue()
}

// PreFlightPhase.Execute executes the pre-flight stub phase before
// reconciliation has been performed.
func (phase *PreFlightPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return r.PreFlight()
}
`
