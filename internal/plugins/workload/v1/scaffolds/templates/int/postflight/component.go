// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package postflight

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

var _ machinery.Template = &Component{}

// Component scaffolds the workload's postflight function.
type Component struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Component) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"postflight",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = componentTemplate

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package postflight

import (
	"{{ .Repo }}/apis/common"
)

// {{ .Resource.Kind }}PostFlight performs postflight logic that happens after
// controller reconciliation is performed.
func {{ .Resource.Kind }}PostFlight(
	reconciler common.ComponentReconciler,
) (ready bool, err error) {
	return true, nil
}
`
