package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Constants{}

// Constants scaffolds controller constants common to controllers.
type Constants struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Constants) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "utils", "constants.go")

	f.TemplateBody = constantsTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const constantsTemplate = `{{ .Boilerplate }}

package utils

const (
	// FieldManager sets the field manager of the controller in the metadata.managedFields
	// specification of a resource.
	FieldManager = "reconciler"

	// optimisticLockErrorMsg is a common message that is returned from the API when an
	// a locking error occurs.  We may safe requeue when we see this message.
	optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"
)
`
