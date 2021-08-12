package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CliCmdRoot{}

// CliCmdRoot scaffolds the root command file for the companion CLI.
type CliCmdRoot struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	// CliRootCmd is the root command for the companion CLI
	CliRootCmd        string
	CliRootCmdVarName string
	// CliRootDescription is the command description given by the CLI help info
	CliRootDescription string
}

func (f *CliCmdRoot) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "root.go")

	f.TemplateBody = cliCmdRootTemplate

	return nil
}

const cliCmdRootTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"
)

// {{ .CliRootCmdVarName }}Command represents the base command when called without any subcommands.
type {{ .CliRootCmdVarName }}Command struct {
	*cobra.Command
}

// New{{ .CliRootCmdVarName }}Command returns an instance of the {{ .CliRootCmdVarName }}Command.
func New{{ .CliRootCmdVarName }}Command() *{{ .CliRootCmdVarName }}Command {
	c := &{{ .CliRootCmdVarName }}Command{
		Command: &cobra.Command{
			Use:   "{{ .CliRootCmd }}",
			Short: "{{ .CliRootDescription }}",
			Long:  "{{ .CliRootDescription }}",
		},
	}

	c.addSubCommands()

	return c
}

// Run represents the main entry point into the command
// This is called by main.main() to execute the root command.
func (c *{{ .CliRootCmdVarName }}Command) Run() {
	cobra.CheckErr(c.Execute())
}

// addSubCommands adds any additional subCommands to the root command.
func (c *{{ .CliRootCmdVarName }}Command) addSubCommands() {
	c.newGenerateCommand()
	c.newInitCommand()
}
`
