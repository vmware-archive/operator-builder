// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package config

import (
	"errors"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
)

var ErrConfigMustExist = errors.New("no workload config provided - workload config required")

// Processor is an object that stores information necessary for generating object source
// code.
type Processor struct {
	Path string

	// Workload represents the top-level configuration (e.g. as passed in via the --workload-config) flag
	// from the command line, while Children represents subordinate configurations that the parent files such
	// as the componentFiles field.
	Workload kinds.Workload
	Children []*Processor

	// The following are used by the scaffolder interface.
	KubebuilderConfig   *config.Config
	KubebuilderResource *resource.Resource
}

// NewProcessor will return a new workload config processor given a path.  An error is returned if the workload config
// does not exist at a path.
func NewProcessor(configPath string) (*Processor, error) {
	if configPath == "" {
		return nil, ErrConfigMustExist
	}

	return &Processor{Path: configPath}, nil
}
