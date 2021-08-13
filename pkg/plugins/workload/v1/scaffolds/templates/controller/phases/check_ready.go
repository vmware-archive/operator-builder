package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CheckReady{}

// CheckReady scaffolds the check ready phase methods.
type CheckReady struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *CheckReady) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "check_ready.go")

	f.TemplateBody = checkReadyTemplate

	return nil
}

const checkReadyTemplate = `{{ .Boilerplate }}

package phases

import (
	common "{{ .Repo }}/apis/common"
)

// CheckReadyPhase.Execute executes checking for a parent components readiness status.
func (phase *CheckReadyPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return r.CheckReady()
}
`
