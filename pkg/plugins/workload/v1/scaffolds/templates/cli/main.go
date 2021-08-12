package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Main{}

// Main scaffolds the main package for the companion CLI.
type Main struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	// CliRootCmd is the root command for the companion CLI
	CliRootCmd        string
	CliRootCmdVarName string
}

func (f *Main) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "main.go")

	f.TemplateBody = cliMainTemplate

	return nil
}

const cliMainTemplate = `{{ .Boilerplate }}

package main

import (
	"{{ .Repo }}/cmd/{{ .CliRootCmd }}/commands"
)

func main() {
	{{ .CliRootCmd }} := commands.New{{ .CliRootCmdVarName }}Command()
	{{ .CliRootCmd }}.Run()
}
`
