package v1

import (
	"fmt"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds"
)

var _ plugin.EditSubcommand = &editSubcommand{}

type editSubcommand struct {
	config config.Config

	multigroup bool
}

func (p *editSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `This command will edit the project configuration.
Features supported:
  - Toggle between single or multi group projects.
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Enable the multigroup layout
  %[1]s edit --multigroup

  # Disable the multigroup layout
  %[1]s edit --multigroup=false
`, cliMeta.CommandName)
}

func (p *editSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&p.multigroup, "multigroup", false, "enable or disable multigroup layout")
}

func (p *editSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	return nil
}

func (p *editSubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewEditScaffolder(p.config, p.multigroup)
	scaffolder.InjectFS(fs)
	return scaffolder.Scaffold()
}
