// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Complete{}

// Complete scaffolds the complete phase methods.
type Complete struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Complete) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "complete.go")

	f.TemplateBody = completeTemplate

	return nil
}

const completeTemplate = `{{ .Boilerplate }}

package phases

import (
	"context"

	"{{ .Repo }}/apis/common"
)

// CompletePhase executes the completion of a reconciliation loop.
func CompletePhase(ctx context.Context, r common.ComponentReconciler) (bool, error) {
	r.GetComponent().SetReadyStatus(true)
	r.GetLogger().V(0).Info("successfully reconciled", "kind", r.GetComponent().GetComponentGVK().Kind)

	return true, nil
}
`
