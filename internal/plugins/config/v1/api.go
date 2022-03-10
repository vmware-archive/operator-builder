// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadconfig "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/config"
)

type createAPISubcommand struct {
	workloadConfigPath string
	processor          *workloadconfig.Processor
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	p.processor = processor

	pluginConfig := workloadconfig.Plugin{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: processor.Workload.GetRootCommand().Name,
	}

	if err := c.EncodePluginConfig(workloadconfig.PluginKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode plugin config at key %s, %w", workloadconfig.PluginKey, err)
	}

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	// set from config file if not provided with command line flag
	if res.Group == "" {
		res.Group = p.processor.Workload.GetAPIGroup()
	}

	if res.Version == "" {
		res.Version = p.processor.Workload.GetAPIVersion()
	}

	if res.Kind == "" {
		res.Kind = p.processor.Workload.GetAPIKind()
		res.Plural = resource.RegularPlural(p.processor.Workload.GetAPIKind())
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
