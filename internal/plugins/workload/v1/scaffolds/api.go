// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package scaffolds

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/api"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/api/resources"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/cli"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/config/crd"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/config/samples"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/controller"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/dependencies"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/mutate"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/test/e2e"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ plugins.Scaffolder = &apiScaffolder{}

var (
	ErrScaffoldAPITypes       = errors.New("error scaffolding api types")
	ErrScaffoldKindInfo       = errors.New("error scaffolding kind information")
	ErrScaffoldResources      = errors.New("error scaffolding resources")
	ErrScaffoldChildResources = errors.New("error scaffolding child resource definitions")
)

type apiScaffolder struct {
	config             config.Config
	resource           *resource.Resource
	boilerplate        string
	boilerplatePath    string
	workload           workloadv1.WorkloadAPIBuilder
	cliRootCommandName string

	fs machinery.Filesystem
}

// NewAPIScaffolder returns a new Scaffolder for project initialization operations.
func NewAPIScaffolder(
	cfg config.Config,
	res *resource.Resource,
	workload workloadv1.WorkloadAPIBuilder,
	cliRootCommandName string,
) plugins.Scaffolder {
	return &apiScaffolder{
		config:             cfg,
		resource:           res,
		boilerplatePath:    "hack/boilerplate.go.txt",
		workload:           workload,
		cliRootCommandName: cliRootCommandName,
	}
}

// InjectFS implements cmdutil.Scaffolder.
func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

//nolint:funlen,gocyclo //this will be refactored later
// scaffold implements cmdutil.Scaffolder.
func (s *apiScaffolder) Scaffold() error {
	log.Println("Building API...")

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return fmt.Errorf("unable to read boilerplate file %s, %w", s.boilerplatePath, err)
	}

	s.boilerplate = string(boilerplate)

	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(s.boilerplate),
		machinery.WithResource(s.resource),
	)

	// scaffold the api
	if err := s.scaffoldAPI(scaffold, s.workload); err != nil {
		return fmt.Errorf("error scaffolding workload apis; %w", err)
	}

	//nolint:nestif //this will be refactored later
	// API types
	if s.workload.IsStandalone() {
		err = scaffold.Execute(
			&templates.MainUpdater{
				WireResource:   true,
				WireController: true,
			},
			&controller.Controller{
				PackageName:       s.workload.GetPackageName(),
				RBACRules:         s.workload.GetRBACRules(),
				OwnershipRules:    s.workload.GetOwnershipRules(),
				HasChildResources: s.workload.HasChildResources(),
				IsStandalone:      s.workload.IsStandalone(),
				IsComponent:       s.workload.IsComponent(),
			},
			&controller.Phases{
				PackageName: s.workload.GetPackageName(),
			},
			&dependencies.Component{},
			&mutate.Component{},
			&samples.CRDSample{
				SpecFields:      s.workload.GetAPISpecFields(),
				IsClusterScoped: s.workload.IsClusterScoped(),
			},
		)
		if err != nil {
			return fmt.Errorf("unable to scaffold standalone workload, %w", err)
		}

		if err := s.scaffoldE2ETests(scaffold, s.workload); err != nil {
			return fmt.Errorf("unable to scaffold standalone workload e2e tests, %w", err)
		}
	} else {
		// collection API
		err = scaffold.Execute(
			&templates.MainUpdater{
				WireResource:   true,
				WireController: true,
			},
			&controller.Controller{
				PackageName:       s.workload.GetPackageName(),
				RBACRules:         s.workload.GetRBACRules(),
				OwnershipRules:    s.workload.GetOwnershipRules(),
				HasChildResources: s.workload.HasChildResources(),
				IsStandalone:      s.workload.IsStandalone(),
				IsComponent:       s.workload.IsComponent(),
			},
			&controller.Phases{
				PackageName: s.workload.GetPackageName(),
			},
			&dependencies.Component{},
			&mutate.Component{},
			&samples.CRDSample{
				SpecFields:      s.workload.GetAPISpecFields(),
				IsClusterScoped: s.workload.IsClusterScoped(),
			},
			&crd.Kustomization{},
		)
		if err != nil {
			return fmt.Errorf("unable to scaffold collection workload, %w", err)
		}

		if err := s.scaffoldE2ETests(scaffold, s.workload); err != nil {
			return fmt.Errorf("unable to scaffold collection workload e2e tests, %w", err)
		}

		for _, component := range s.workload.GetComponents() {
			componentScaffold := machinery.NewScaffold(s.fs,
				machinery.WithConfig(s.config),
				machinery.WithBoilerplate(string(boilerplate)),
				machinery.WithResource(component.GetComponentResource(
					s.config.GetDomain(),
					s.config.GetRepository(),
					component.IsClusterScoped(),
				)),
			)

			err = componentScaffold.Execute(
				&templates.MainUpdater{
					WireResource:   true,
					WireController: true,
				},
				&controller.Controller{
					PackageName:       component.GetPackageName(),
					RBACRules:         component.GetRBACRules(),
					OwnershipRules:    component.GetOwnershipRules(),
					HasChildResources: component.HasChildResources(),
					IsStandalone:      component.IsStandalone(),
					IsComponent:       component.IsComponent(),
					Collection:        s.workload.(*workloadv1.WorkloadCollection),
				},
				&controller.Phases{
					PackageName: s.workload.GetPackageName(),
				},
				&dependencies.Component{},
				&mutate.Component{},
				&samples.CRDSample{
					SpecFields:      component.Spec.APISpecFields,
					IsClusterScoped: s.workload.IsClusterScoped(),
				},
				&crd.Kustomization{},
			)
			if err != nil {
				return fmt.Errorf("unable to scaffold component workload %s, %w", component.Name, err)
			}

			if err := s.scaffoldE2ETests(componentScaffold, component); err != nil {
				return fmt.Errorf("unable to scaffold component workload e2e tests, %w", err)
			}
		}
	}

	// scaffold the companion CLI last only if we have a root command name
	if s.cliRootCommandName != "" {
		if err = s.scaffoldCLI(scaffold); err != nil {
			return fmt.Errorf("error scaffolding CLI; %w", err)
		}
	}

	return nil
}

// scaffoldAPI runs the specific logic to scaffold anything existing in the apis directory.
func (s *apiScaffolder) scaffoldAPI(
	scaffold *machinery.Scaffold,
	workload workloadv1.WorkloadAPIBuilder,
) error {
	// override the scaffold if we have a component
	if workload.IsComponent() {
		scaffold = machinery.NewScaffold(s.fs,
			machinery.WithConfig(s.config),
			machinery.WithBoilerplate(s.boilerplate),
			machinery.WithResource(workload.GetComponentResource(
				s.config.GetDomain(),
				s.config.GetRepository(),
				workload.IsClusterScoped(),
			)),
		)
	}

	// scaffold the base api types
	if err := scaffold.Execute(
		&api.Types{Builder: workload},
		&api.Group{},
	); err != nil {
		return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldAPITypes, workload)
	}

	// scaffold the specific kind files
	if err := scaffold.Execute(
		&api.Kind{},
		&api.KindLatest{PackageName: workload.GetPackageName()},
		&api.KindUpdater{},
	); err != nil {
		return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldKindInfo, workload)
	}

	// scaffold the resources
	if err := scaffold.Execute(
		&resources.Resources{Builder: workload},
	); err != nil {
		return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldResources, workload)
	}

	// child resource definition files
	// these are the resources defined in the static yaml manifests
	for _, sourceFile := range *workload.GetSourceFiles() {
		if err := scaffold.Execute(
			&resources.Definition{Builder: workload, SourceFile: sourceFile},
		); err != nil {
			return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldChildResources, workload)
		}
	}

	// scaffold the components
	if workload.IsCollection() {
		for _, component := range workload.GetComponents() {
			if err := s.scaffoldAPI(scaffold, component); err != nil {
				return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldResources, component)
			}
		}
	}

	return nil
}

// scaffoldCLI runs the specific logic to scaffold the companion CLI.
func (s *apiScaffolder) scaffoldCLI(scaffold *machinery.Scaffold) error {
	// obtain a list of workload commands to generate, to include the parent collection
	// and its children
	workloadCommands := make([]workloadv1.WorkloadAPIBuilder, len(s.workload.GetComponents())+1)
	workloadCommands[0] = s.workload

	if len(s.workload.GetComponents()) > 0 {
		for i, component := range s.workload.GetComponents() {
			workloadCommands[i+1] = component
		}
	}

	for _, workloadCommand := range workloadCommands {
		// create this component as a kubebuilder component resource for those
		// commands that need it
		componentResource := workloadCommand.GetComponentResource(
			s.config.GetDomain(),
			s.config.GetRepository(),
			workloadCommand.IsClusterScoped(),
		)

		// scaffold init subcommand
		if err := scaffold.Execute(
			&cli.CmdInitSub{Builder: workloadCommand, ComponentResource: componentResource},
			&cli.CmdInitSubUpdater{Builder: workloadCommand, ComponentResource: componentResource},
		); err != nil {
			return fmt.Errorf("unable to scaffold init subcommand, %w", err)
		}

		// scaffold the generate command unless we have a collection without resources
		if (workloadCommand.HasChildResources() && workloadCommand.IsCollection()) || !workloadCommand.IsCollection() {
			if err := scaffold.Execute(
				&cli.CmdGenerateSub{Builder: workloadCommand, ComponentResource: componentResource},
				&cli.CmdGenerateSubUpdater{Builder: workloadCommand, ComponentResource: componentResource},
			); err != nil {
				return fmt.Errorf("unable to scaffold generate subcommand, %w", err)
			}
		}

		// scaffold version subcommand
		if err := scaffold.Execute(
			&cli.CmdVersionSub{Builder: workloadCommand, ComponentResource: componentResource},
			&cli.CmdVersionSubUpdater{Builder: workloadCommand, ComponentResource: componentResource},
		); err != nil {
			return fmt.Errorf("unable to scaffold version subcommand, %w", err)
		}

		// scaffold the root command
		if err := scaffold.Execute(
			&cli.CmdRootUpdater{
				InitCommand:     true,
				GenerateCommand: true,
				VersionCommand:  true,
				Builder:         workloadCommand,
			},
		); err != nil {
			return fmt.Errorf("error updating root.go, %w", err)
		}
	}

	return nil
}

// scaffoldE2ETests run the specific logic to scaffold the end to end tests.
func (s *apiScaffolder) scaffoldE2ETests(
	scaffold *machinery.Scaffold,
	workload workloadv1.WorkloadAPIBuilder,
) error {
	if err := scaffold.Execute(&e2e.WorkloadTest{Builder: workload}); err != nil {
		return fmt.Errorf("error updating test/e2e/%s_%s_%s_test.go; %w",
			workload.GetAPIGroup(), workload.GetAPIVersion(), strings.ToLower(workload.GetAPIKind()),
			err,
		)
	}

	return nil
}
