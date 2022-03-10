// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/commands/subcommand"
	workloadconfig "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/config"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
)

type createAPISubcommand struct {
	config config.Config

	resource *resource.Resource

	workloadConfigPath string
	cliRootCommandName string
	workload           kinds.Workload
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Build a new API that can capture state for workloads
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add API attributes defined by a workload config file
  %[1]s create api --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	p.config = c

	var pluginConfig workloadconfig.Plugin
	if err := c.DecodePluginConfig(workloadconfig.PluginKey, &pluginConfig); err != nil {
		return fmt.Errorf("unable to decode operatorbuilder config key at %s, %w", p.workloadConfigPath, err)
	}

	p.workloadConfigPath = pluginConfig.WorkloadConfigPath
	p.cliRootCommandName = pluginConfig.CliRootCommandName

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	p.resource = res

	return nil
}

func (p *createAPISubcommand) PreScaffold(machinery.Filesystem) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	if err := subcommand.CreateAPI(processor); err != nil {
		return err
	}

	p.workload = processor.Workload

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewAPIScaffolder(
		p.config,
		p.resource,
		p.workload,
		p.cliRootCommandName,
	)
	scaffolder.InjectFS(fs)

	if err := scaffolder.Scaffold(); err != nil {
		return fmt.Errorf("unable to scaffold api, %w", err)
	}

	return nil
}
