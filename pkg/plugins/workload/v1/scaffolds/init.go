package scaffolds

import (
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/cli"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

const CobraVersion = "v1.1.3"

const (
	// ControllerRuntimeVersion is the kubernetes-sigs/controller-runtime version to be used in the project
	ControllerRuntimeVersion = "v0.7.2"
	// ControllerToolsVersion is the kubernetes-sigs/controller-tools version to be used in the project
	ControllerToolsVersion = "v0.6.1"
	// KustomizeVersion is the kubernetes-sigs/kustomize version to be used in the project
	KustomizeVersion = "v3.8.7"
	// CobraVersion = "v1.1.3"

	imageName = "controller:latest"
)

var _ plugins.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	config          config.Config
	boilerplatePath string
	workload        workloadv1.WorkloadInitializer

	fs machinery.Filesystem
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations.
func NewInitScaffolder(cfg config.Config, workload workloadv1.WorkloadInitializer) plugins.Scaffolder {
	return &initScaffolder{
		config:          cfg,
		boilerplatePath: "hack/boilerplate.go.txt",
		workload:        workload,
	}
}

// InjectFS implements cmdutil.Scaffolder.
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder.
func (s *initScaffolder) Scaffold() error {
	fmt.Println("Adding workload scaffolding...")

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
	)

	if s.workload.HasRootCmdName() {
		err = scaffold.Execute(
			&cli.Main{
				RootCmd:        s.workload.GetRootCmdName(),
				RootCmdVarName: utils.ToPascalCase(s.workload.GetRootCmdName()),
			},
			&cli.CmdRoot{
				RootCmd:            s.workload.GetRootCmdName(),
				RootCmdVarName:     utils.ToPascalCase(s.workload.GetRootCmdName()),
				RootCmdDescription: s.workload.GetRootCmdDescr(),
			},
			&templates.Makefile{
				RootCmd: s.workload.GetRootCmdName(),
			},
			&templates.Project{
				RootCmd: s.workload.GetRootCmdName(),
			},
		)
		if err != nil {
			return err
		}
	}

	err = scaffold.Execute(
		&templates.Main{},
		&templates.GoMod{
			ControllerRuntimeVersion: scaffolds.ControllerRuntimeVersion,
			CobraVersion:             CobraVersion,
		},
		&templates.Dockerfile{},
	)
	if err != nil {
		return err
	}

	return nil
}
