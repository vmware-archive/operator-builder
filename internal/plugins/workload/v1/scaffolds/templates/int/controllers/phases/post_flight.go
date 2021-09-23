// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PostFlight{}

// PostFlight scaffolds the post-flight phase methods.
type PostFlight struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *PostFlight) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "post_flight.go")

	f.TemplateBody = postFlightTemplate

	return nil
}

const postFlightTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// PostFlightPhase.DefaultRequeue defines the default requeue result for this
// phase.
func (phase *PostFlightPhase) DefaultRequeue() ctrl.Result {
	return Requeue()
}

// PostFlightPhase.Execute executes the post-flight stub phase after
// reconciliation has been performed.
func (phase *PostFlightPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return r.PostFlight()
}
`
