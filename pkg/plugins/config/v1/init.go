package v1

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

type initSubcommand struct {
	workloadConfigPath string
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	workload, err := workloadv1.ProcessInitConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return err
	}

	pluginConfig := workloadv1.PluginConfig{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: workload.GetRootCmdName(),
	}

	if err := c.EncodePluginConfig(workloadv1.PluginConfigKey, pluginConfig); err != nil {
		return err
	}

	if err := c.SetDomain(workload.GetDomain()); err != nil {
		return err
	}

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
