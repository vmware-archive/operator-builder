package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &CmdInitSub{}

// CmdInitSub scaffolds the companion CLI's init subcommand for the
// workload.  This where the actual init logic lives.
type CmdInitSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	CliRootCmd        string
	CliRootCmdVarName string
	CmdCmdName        string
	CmdCmdDescr       string
	CmdCmdVarName     string
	CmdCmdFileName    string
	SpecFields        *[]workloadv1.APISpecField
	IsComponent       bool
	ComponentResource *resource.Resource

	InitCommandName  string
	InitCommandDescr string
}

func (f *CmdInitSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Path = filepath.Join(
			"cmd", f.CliRootCmd, "commands",
			fmt.Sprintf("%s_init.go", f.CmdCmdFileName),
		)
		f.Resource = f.ComponentResource
	} else {
		f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "init.go")
	}

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	f.TemplateBody = cliCmdInitSubTemplate

	return nil
}

const cliCmdInitSubTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultManifest{{ .CmdCmdVarName }} = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
spec:
{{- range .SpecFields }}
  {{ .SampleField -}}
{{ end }}
` + "`" + `

{{ if not .IsComponent -}}
// newInitCommand creates a new instance of the init subcommand.
func (c *{{ .CliRootCmdVarName }}Command) newInitCommand() {
{{- else }}
// newInit{{ .CmdCmdVarName }}Command creates a new instance of the  init{{ .CmdCmdVarName }} subcommand.
func (i *initCommand) newInit{{ .CmdCmdVarName }}Command() {
{{- end }}
	init{{ .CmdCmdVarName }}Cmd := &cobra.Command{
		{{ if .IsComponent -}}
		Use:   "{{ .CmdCmdName }}",
		Short: "{{ .CmdCmdDescr }}",
		Long: "{{ .CmdCmdDescr }}",
		{{- else -}}
		Use:   "{{ .InitCommandName }}",
		Short: "{{ .InitCommandDescr }}",
		Long: "{{ .InitCommandDescr }}",
		{{- end }}
		RunE: func(cmd *cobra.Command, args []string) error {
			outputStream := os.Stdout

			if _, err := outputStream.WriteString(defaultManifest{{ .CmdCmdVarName }}); err != nil {
				return fmt.Errorf("failed to write outout, %w", err)
			}

			return nil
		},
	}

	{{ if .IsComponent -}}
	i.AddCommand(init{{ .CmdCmdVarName }}Cmd)
	{{- else -}}
	c.AddCommand(init{{ .CmdCmdVarName }}Cmd)
	{{- end -}}
}
`
