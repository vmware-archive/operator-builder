package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Errors{}

// Errors scaffolds controller errors common to controllers.
type Errors struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Errors) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "utils", "errors.go")

	f.TemplateBody = errorsTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const errorsTemplate = `{{ .Boilerplate }}

package utils

import (
	"strings"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

// IsOptimisticLockError checks to see if the error is a locking error.
func IsOptimisticLockError(err error) bool {
	return strings.Contains(err.Error(), optimisticLockErrorMsg)
}

// IgnoreNotFound ignores an error when the error indicates that an item
// is not found while calling the Kubernetes API.
func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}

	return err
}
`
