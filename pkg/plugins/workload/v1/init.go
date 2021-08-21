package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"

	"github.com/spf13/pflag"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

type initSubcommand struct {
	config config.Config
	// For help text.
	commandName string

	// boilerplate options
	license string
	owner   string

	// go config options
	repo string

	// flags
	fetchDeps bool

	// operator-builder specific options
	workloadConfigPath string
	workload           workloadv1.WorkloadInitializer
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Add workload management scaffolding to a new project
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add project scaffolding defined by a workload config file
  %[1]s init --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	// dependency args
	fs.BoolVar(&p.fetchDeps, "fetch-deps", true, "ensure dependencies are downloaded")

	// boilerplate args
	fs.StringVar(&p.license, "license", "apache2",
		"license to use to boilerplate, may be one of 'apache2', 'none'")
	fs.StringVar(&p.owner, "owner", "", "owner to add to the copyright")

	// project args
	fs.StringVar(&p.repo, "repo", "", "name to use for go module (e.g., github.com/user/repo), "+
		"defaults to the go package of the current working directory.")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	p.commandName = CommandName

	// Try to guess repository if flag is not set.
	if p.repo == "" {
		repoPath, err := golang.FindCurrentRepo()
		if err != nil {
			return fmt.Errorf("error finding current repository: %v", err)
		}
		p.repo = repoPath
	}
	if err := p.config.SetRepository(p.repo); err != nil {
		return err
	}

	// operator builder always uses multi-group APIs
	if err := c.SetMultiGroup(); err != nil {
		return err
	}

	var taxi workloadv1.ConfigTaxi
	if err := c.DecodePluginConfig(workloadv1.ConfigTaxiKey, &taxi); err != nil {
		return err
	}

	p.workloadConfigPath = taxi.WorkloadConfigPath

	return nil
}

func (p *initSubcommand) PreScaffold(machinery.Filesystem) error {
	// Check if the current directory has not files or directories which does not allow to init the project
	if err := checkDir(); err != nil {
		return err
	}

	// load the workload config
	workload, err := workloadv1.ProcessInitConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return err
	}

	// validate the workload config
	if err := workload.Validate(); err != nil {
		return err
	}

	p.workload = workload

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewInitScaffolder(p.config, p.workload)
	scaffolder.InjectFS(fs)

	if err := scaffolder.Scaffold(); err != nil {
		return err
	}

	if !p.fetchDeps {
		fmt.Println("Skipping fetching dependencies.")
		return nil
	}

	// Ensure that we are pinning controller-runtime version
	// xref: https://github.com/kubernetes-sigs/kubebuilder/issues/997
	err := util.RunCmd("Get controller runtime", "go", "get",
		"sigs.k8s.io/controller-runtime@"+scaffolds.ControllerRuntimeVersion)
	if err != nil {
		return err
	}

	return nil
}

func (p *initSubcommand) PostScaffold() error {
	err := util.RunCmd("Update dependencies", "go", "mod", "tidy")
	if err != nil {
		return err
	}

	fmt.Printf("Next: define a resource with:\n$ %s create api\n", p.commandName)
	return nil
}

// checkDir will return error if the current directory has files which are not allowed.
// Note that, it is expected that the directory to scaffold the project is cleaned.
// Otherwise, it might face issues to do the scaffold.
func checkDir() error {
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Allow directory trees starting with '.'
			if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			// Allow files starting with '.'
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			// Allow files ending with '.md' extension
			if strings.HasSuffix(info.Name(), ".md") && !info.IsDir() {
				return nil
			}
			// Allow capitalized files except PROJECT
			isCapitalized := true
			for _, l := range info.Name() {
				if !unicode.IsUpper(l) {
					isCapitalized = false
					break
				}
			}
			if isCapitalized && info.Name() != "PROJECT" {
				return nil
			}
			// Allow files in the following list
			allowedFiles := []string{
				"go.mod", // user might run `go mod init` instead of providing the `--flag` at init
				"go.sum", // auto-generated file related to go.mod
			}
			for _, allowedFile := range allowedFiles {
				if info.Name() == allowedFile {
					return nil
				}
			}
			// Do not allow any other file
			return fmt.Errorf(
				"target directory is not empty (only %s, files and directories with the prefix \".\", "+
					"files with the suffix \".md\" or capitalized files name are allowed); "+
					"found existing file %q", strings.Join(allowedFiles, ", "), path)
		})
	if err != nil {
		return err
	}
	return nil
}
