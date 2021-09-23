// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package postcreate

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

var _ machinery.Template = &Component{}

// Component scaffolds the workload's postcreate function.
type Component struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Component) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"postcreate",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = componentTemplate

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package postcreate

import (
	"{{ .Repo }}/apis/common"
)

// {{ .Resource.Kind }}PostCreate performs postcreate logic that happens after
// resource creation is performed.
func {{ .Resource.Kind }}PostCreate(
	reconciler common.ComponentReconciler,
) (ready bool, err error) {
	return true, nil
}
`
