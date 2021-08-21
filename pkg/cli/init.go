package cli

import (
	kbcli "sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"

	configv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/config/v1"
	licensev1 "github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/license/v1"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1"
)

var version = "unstable"

func NewKubebuilderCLI(commandName string) (*kbcli.CLI, error) {
	gov3Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		licensev1.Plugin{},
		kustomizecommonv1.Plugin{},
		configv1.Plugin{},
		workloadv1.Plugin{},
	)

	c, err := kbcli.New(
		kbcli.WithCommandName(commandName),
		kbcli.WithVersion(version),
		kbcli.WithPlugins(
			gov3Bundle,
			&licensev1.Plugin{},
			&kustomizecommonv1.Plugin{},
			&declarativev1.Plugin{},
			&workloadv1.Plugin{},
		),
		kbcli.WithDefaultPlugins(cfgv3.Version, gov3Bundle),
		kbcli.WithDefaultProjectVersion(cfgv3.Version),
		kbcli.WithExtraCommands(NewUpdateCmd()),
		kbcli.WithCompletion(),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}
