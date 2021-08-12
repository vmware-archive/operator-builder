package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &CliCmdInitSub{}

// CliCmdInitSub scaffolds the companion CLI's init subcommand for the
// workload.  This where the actual init logic lives.
type CliCmdInitSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	CliRootCmd        string
	CliSubCmdName     string
	CliSubCmdDescr    string
	CliSubCmdVarName  string
	CliSubCmdFileName string
	SpecFields        *[]workloadv1.APISpecField
	IsComponent       bool
	ComponentResource *resource.Resource

	InitCommandName  string
	InitCommandDescr string
}

func (f *CliCmdInitSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Path = filepath.Join(
			"cmd", f.CliRootCmd, "commands",
			fmt.Sprintf("%s_init.go", f.CliSubCmdFileName),
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

const defaultManifest{{ .CliSubCmdVarName }} = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
spec:
{{- range .SpecFields }}
  {{ .SampleField -}}
{{ end }}
` + "`" + `

{{ if not .IsComponent -}}
type initCommand struct {
	*cobra.Command
}
{{- end }}

{{ if not .IsComponent -}}
// newInitCommand creates a new instance of the init subcommand.
func (c *{{ .CliRootCmd }}Command) newInitCommand() {
{{- else }}
// new{{ .CliSubCmdVarName }}InitCommand creates a new instance of the {{ .CliSubCmdVarName }} init subcommand.
func (i *initCommand) new{{ .CliSubCmdVarName }}InitCommand() {
{{- end }}
	{{ .CliSubCmdVarName }}InitCmd := &cobra.Command{
		{{ if .IsComponent -}}
		Use:   "{{ .CliSubCmdName }}",
		Short: "{{ .CliSubCmdDescr }}",
		Long: "{{ .CliSubCmdDescr }}",
		{{- else -}}
		Use:   "{{ .InitCommandName }}",
		Short: "{{ .InitCommandDescr }}",
		Long: "{{ .InitCommandDescr }}",
		{{- end }}
		RunE: func(cmd *cobra.Command, args []string) error {
			outputStream := os.Stdout

			if _, err := outputStream.WriteString(defaultManifest{{ .CliSubCmdVarName }}); err != nil {
				return fmt.Errorf("failed to write outout, %w", err)
			}

			return nil
		},
	}

	{{ if .IsComponent -}}
	i.AddCommand({{ .CliSubCmdVarName }}InitCmd)
	{{- else -}}
	c.AddCommand({{ .CliSubCmdVarName }}InitCmd)
	{{- end -}}
}
`
