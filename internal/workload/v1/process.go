package v1

import (
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

func formatProcessError(manifestFile string, err error) error {
	return fmt.Errorf("error processing file %s; %w", manifestFile, err)
}

func determineSourceFileName(manifestFile string) SourceFile {
	var sourceFile SourceFile
	sourceFile.Filename = filepath.Base(manifestFile)                // get filename from path
	sourceFile.Filename = strings.Split(sourceFile.Filename, ".")[0] // strip ".yaml"
	sourceFile.Filename += ".go"                                     // add correct file ext
	sourceFile.Filename = utils.ToFileName(sourceFile.Filename)      // kebab-case to snake_case

	return sourceFile
}

func generateUniqueResourceName(object unstructured.Unstructured) string {
	resourceName := strings.Replace(strings.Title(object.GetName()), "-", "", -1)
	resourceName = strings.Replace(resourceName, ".", "", -1)
	resourceName = strings.Replace(resourceName, ":", "", -1)
	resourceName = fmt.Sprintf("%s%s", object.GetKind(), resourceName)

	return resourceName
}
