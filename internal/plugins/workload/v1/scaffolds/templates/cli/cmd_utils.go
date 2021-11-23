// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CmdUtils{}

// CmdUtils scaffolds the companion CLI's common utility code for the
// workload.  This where the generic logic for a companion CLI lives.
type CmdUtils struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	RootCmdName string
}

func (f *CmdUtils) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmdName, "commands", "utils", "utils.go")

	f.TemplateBody = cliCmdUtilsTemplate

	return nil
}

const cliCmdUtilsTemplate = `{{ .Boilerplate }}

package utils

import (
	"errors"
	"fmt"

	"{{ .Repo }}/apis/common"
)

var ErrInvalidResource = errors.New("supplied resource is incorrect")

// ValidateWorkload validates the unmarshaled version of the workload resource
// manifest.
func ValidateWorkload(workload common.Component) error {
	defaultWorkloadGVK := workload.GetComponentGVK()

	if defaultWorkloadGVK != workload.GetObjectKind().GroupVersionKind() {
		return fmt.Errorf(
			"%w, expected resource of kind: '%s', with group '%s' and version '%s'; "+
				"found resource of kind '%s', with group '%s' and version '%s'",
			ErrInvalidResource,
			defaultWorkloadGVK.Kind,
			defaultWorkloadGVK.Group,
			defaultWorkloadGVK.Version,
			workload.GetObjectKind().GroupVersionKind().Kind,
			workload.GetObjectKind().GroupVersionKind().Group,
			workload.GetObjectKind().GroupVersionKind().Version,
		)
	}

	return nil
}
`
