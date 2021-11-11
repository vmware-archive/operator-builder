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
	"context"

	"{{ .Repo }}/apis/common"
)

// PreFlightPhase executes pre-flight and fail-fast conditions prior to attempting resource creation.
func PreFlightPhase(ctx context.Context, r common.ComponentReconciler) (bool, error) {
	return true, nil
}
`
