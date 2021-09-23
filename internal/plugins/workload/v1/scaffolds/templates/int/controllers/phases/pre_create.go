// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PreCreate{}

// PreCreate scaffolds the pre-create phase methods.
type PreCreate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *PreCreate) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "pre_create.go")

	f.TemplateBody = preCreateTemplate

	return nil
}

const preCreateTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// PreCreatePhase.DefaultRequeue defines the default requeue result for this
// phase.
func (phase *PreCreatePhase) DefaultRequeue() ctrl.Result {
	return Requeue()
}

// PreCreatePhase.Execute executes the pre-create stub phase before
// resource creation has been performed.
func (phase *PreCreatePhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return r.PreCreate()
}
`
