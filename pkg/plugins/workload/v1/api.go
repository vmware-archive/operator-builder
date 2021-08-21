package v1

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	goPlugin "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"

	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

const (
	// defaultCRDVersion is the default CRD API version to scaffold.
	defaultCRDVersion = "v1"
)

// DefaultMainPath is default file path of main.go
const DefaultMainPath = "main.go"

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

type createAPISubcommand struct {
	config config.Config

	options *goPlugin.Options

	resource *resource.Resource

	// Check if we have to scaffold resource and/or controller
	resourceFlag   *pflag.Flag
	controllerFlag *pflag.Flag

	// force indicates that the resource should be created even if it already exists
	force bool

	// runMake indicates whether to run make or not after scaffolding APIs
	runMake bool

	// operator-builder specific options
	workloadConfigPath string
	workload           workloadv1.WorkloadAPIBuilder
	project            workloadv1.Project
}

func (p *createAPISubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Build a new API that can capture state for workloads
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add API attributes defined by a workload config file
  %[1]s create api --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	p.config = c

	var taxi workloadv1.ConfigTaxi
	if err := c.DecodePluginConfig(workloadv1.ConfigTaxiKey, &taxi); err != nil {
		return err
	}

	p.workloadConfigPath = taxi.WorkloadConfigPath

	return nil
}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&p.runMake, "make", true, "if true, run `make generate` after generating files")

	fs.BoolVar(&p.force, "force", false,
		"attempt to create resource even if it already exists")

	p.options = &goPlugin.Options{}

	fs.StringVar(&p.options.Plural, "plural", "", "resource irregular plural form")

	fs.BoolVar(&p.options.DoAPI, "resource", true,
		"if set, generate the resource without prompting the user")
	p.resourceFlag = fs.Lookup("resource")
	fs.StringVar(&p.options.CRDVersion, "crd-version", defaultCRDVersion,
		"version of CustomResourceDefinition to scaffold. Options: [v1, v1beta1]")
	fs.BoolVar(&p.options.Namespaced, "namespaced", true, "resource is namespaced")

	fs.BoolVar(&p.options.DoController, "controller", true,
		"if set, generate the controller without prompting the user")
	p.controllerFlag = fs.Lookup("controller")
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	p.resource = res

	// TODO: re-evaluate whether y/n input still makes sense. We should probably always
	//       scaffold the resource and controller.
	// Ask for API and Controller if not specified
	reader := bufio.NewReader(os.Stdin)
	if !p.resourceFlag.Changed {
		fmt.Println("Create Resource [y/n]")
		p.options.DoAPI = util.YesNo(reader)
	}
	if !p.controllerFlag.Changed {
		fmt.Println("Create Controller [y/n]")
		p.options.DoController = util.YesNo(reader)
	}

	p.options.UpdateResource(p.resource, p.config)

	if err := p.resource.Validate(); err != nil {
		return err
	}

	// In case we want to scaffold a resource API we need to do some checks
	if p.options.DoAPI {
		// Check that resource doesn't have the API scaffolded or flag force was set
		if r, err := p.config.GetResource(p.resource.GVK); err == nil && r.HasAPI() && !p.force {
			return errors.New("API resource already exists")
		}

		// Check that the provided group can be added to the project
		if !p.config.IsMultiGroup() && p.config.ResourcesLength() != 0 && !p.config.HasGroup(p.resource.Group) {
			return fmt.Errorf("multiple groups are not allowed by default, " +
				"to enable multi-group visit https://kubebuilder.io/migration/multi-group.html")
		}

		// Check CRDVersion against all other CRDVersions in p.config for compatibility.
		if util.HasDifferentCRDVersion(p.config, p.resource.API.CRDVersion) {
			return fmt.Errorf("only one CRD version can be used for all resources, cannot add %q",
				p.resource.API.CRDVersion)
		}
	}

	return nil
}

func (p *createAPISubcommand) PreScaffold(machinery.Filesystem) error {
	// check if main.go is present in the root directory
	if _, err := os.Stat(DefaultMainPath); os.IsNotExist(err) {
		return fmt.Errorf("%s file should present in the root directory", DefaultMainPath)
	}

	// load the workload config
	workload, err := workloadv1.ProcessAPIConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return err
	}

	// validate the workload config
	err = workload.Validate()
	if err != nil {
		return err
	}

	p.workload = workload

	// get WORKLOAD project config file
	projectFile, err := ioutil.ReadFile("WORKLOAD")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(projectFile, &p.project)
	if err != nil {
		return err
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewAPIScaffolder(
		p.config,
		*p.resource,
		p.workload,
		&p.project,
	)
	scaffolder.InjectFS(fs)

	return scaffolder.Scaffold()
}

func (p *createAPISubcommand) PostScaffold() error {
	err := util.RunCmd("Update dependencies", "go", "mod", "tidy")
	if err != nil {
		return err
	}
	if p.runMake && p.resource.HasAPI() {
		err = util.RunCmd("Running make", "make", "generate")
		if err != nil {
			return err
		}
		fmt.Print("Next: implement your new API and generate the manifests (e.g. CRDs,CRs) with:\n$ make manifests\n")
	}

	return nil
}
