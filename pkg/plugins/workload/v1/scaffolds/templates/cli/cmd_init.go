package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

const (
	initCommandName  = "init"
	initCommandDescr = "Write a sample custom resource manifest for a workload to standard out"
)

var _ machinery.Template = &CliCmdInit{}

// CliCmdInit scaffolds the companion CLI's init subcommand for
// component workloads.  The init logic will live in the workload's
// subcommand to this command; see cmd_init_sub.go.
type CliCmdInit struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	CliRootCmd string

	InitCommandName  string
	InitCommandDescr string

	Collection *workloadv1.WorkloadCollection
}

func (f *CliCmdInit) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "init.go")

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	f.TemplateBody = cliCmdInitTemplate

	return nil
}

const cliCmdInitTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"
)

type initCommand struct {
	*cobra.Command
}

// newInitCommand creates a new instance of the init subcommand.
func (c *{{ .CliRootCmd }}Command) newInitCommand() {
	initCmd := &initCommand{}

	initCmd.Command = &cobra.Command{
		Use:   "{{ .InitCommandName }}",
		Short: "{{ .InitCommandDescr }}",
		Long: "{{ .InitCommandDescr }}",
	}

	initCmd.addCommands()
	c.AddCommand(initCmd.Command)
}

func (i *initCommand) addCommands() {
	{{- range $component := .Collection.Spec.Components }}
	i.new{{ $component.Spec.CompanionCliSubcmd.VarName }}InitCommand()
	{{- end }}
}
`
