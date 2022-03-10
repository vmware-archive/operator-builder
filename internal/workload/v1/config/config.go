// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package config

import "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"

const PluginKey = "operatorBuilder"

// Plugin contains the project config values which are stored in the
// PROJECT file under plugins.operatorBuilder.
type Plugin struct {
	WorkloadConfigPath string `json:"workloadConfigPath" yaml:"workloadConfigPath"`
	CliRootCommandName string `json:"cliRootCommandName" yaml:"cliRootCommandName"`
}

// GetWorkloads gets all of the workloads for a config processor in a flattened
// fashion.
func GetWorkloads(processor *Processor) []kinds.Workload {
	workloads := []kinds.Workload{processor.Workload}

	// return the array with the single workload if we have no children
	if len(processor.Children) == 0 {
		return workloads
	}

	// get the workloads for a child and append it to the array
	for _, child := range processor.Children {
		workloads = append(workloads, GetWorkloads(child)...)
	}

	return workloads
}

// GetProcessors gets all of the processors to include the parent and children processors.
func GetProcessors(processor *Processor) []*Processor {
	processors := []*Processor{processor}

	// return array with single processor if we have no children
	if len(processor.Children) == 0 {
		return processors
	}

	// get the processors for a child and append it to the array
	for _, child := range processor.Children {
		processors = append(processors, GetProcessors(child)...)
	}

	return processors
}
