// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PostCreate{}

// PostCreate scaffolds the post-create phase methods.
type PostCreate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *PostCreate) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "post_create.go")

	f.TemplateBody = postCreateTemplate

	return nil
}

const postCreateTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// PostCreatePhase.DefaultRequeue defines the default requeue result for this
// phase.
func (phase *PostCreatePhase) DefaultRequeue() ctrl.Result {
	return Requeue()
}

// PostCreatePhase.Execute executes the post-create stub phase after
// resource creation has been performed.
func (phase *PostCreatePhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return r.PostCreate()
}
`
