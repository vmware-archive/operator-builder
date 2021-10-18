package v1

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

func processManifests(
	workloadPath string,
	resources []string,
	collection bool,
	collectionResources bool,
) (*SourceCode, error) {
	results := &SourceCode{
		SourceFiles:         new([]SourceFile),
		RBACRules:           new(RBACRules),
		OwnershipRules:      new(OwnershipRules),
		collection:          collection,
		collectionResources: collectionResources,
	}

	for _, manifestFile := range resources {
		// capture entire resource manifest file content
		manifests, err := results.processMarkers(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return nil, err
		}

		if collection && !collectionResources {
			continue
		}

		// determine sourceFile filename
		sourceFile := determineSourceFileName(manifestFile)

		var childResources []ChildResource

		for _, manifest := range manifests {
			// decode manifest into unstructured data type
			var manifestObject unstructured.Unstructured

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			err := runtime.DecodeInto(decoder, []byte(manifest), &manifestObject)
			if err != nil {
				return nil, formatProcessError(manifestFile, err)
			}

			// generate a unique name for the resource using the kind and name
			resourceUniqueName := generateUniqueResourceName(manifestObject)
			// determine resource group and version
			resourceVersion, resourceGroup := versionGroupFromAPIVersion(manifestObject.GetAPIVersion())

			// determine group and resource for RBAC rule generation
			results.RBACRules.addRulesForManifest(manifestObject.GetKind(), resourceGroup, manifestObject.Object)

			results.OwnershipRules.addOrUpdateOwnership(
				manifestObject.GetAPIVersion(),
				manifestObject.GetKind(),
				resourceGroup,
			)

			resource := ChildResource{
				Name:       manifestObject.GetName(),
				UniqueName: resourceUniqueName,
				Group:      resourceGroup,
				Version:    resourceVersion,
				Kind:       manifestObject.GetKind(),
			}

			// generate the object source code
			resourceDefinition, err := generate.Generate([]byte(manifest), "resourceObj")
			if err != nil {
				return nil, formatProcessError(manifestFile, err)
			}

			// add the source code to the resource
			resource.SourceCode = resourceDefinition
			resource.StaticContent = manifest

			childResources = append(childResources, resource)
		}

		sourceFile.Children = childResources
		*results.SourceFiles = append(*results.SourceFiles, sourceFile)
	}

	// ensure no duplicate file names exist within the source files
	deduplicateFileNames(results)

	return results, nil
}

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

// deduplicateFileNames dedeplicates the names of the files.  This is because
// we cannot guarantee that files exist in different directories and may have
// naming collisions.
func deduplicateFileNames(templateData *SourceCode) {
	// create a slice to track existing fileNames and preallocate an existing
	// known conflict
	fileNames := make([]string, len(*templateData.SourceFiles)+1)
	fileNames[len(fileNames)-1] = "resources.go"

	// dereference the sourcefiles
	sourceFiles := *templateData.SourceFiles

	for i, sourceFile := range sourceFiles {
		var count int

		for _, fileName := range fileNames {
			if fileName == "" {
				continue
			}

			if sourceFile.Filename == fileName {
				// increase the count which serves as an index to append
				count++

				// adjust the filename
				fields := strings.Split(sourceFile.Filename, ".go")
				sourceFiles[i].Filename = fmt.Sprintf("%s_%v.go", fields[0], count)
			}
		}

		fileNames[i] = sourceFile.Filename
	}
}
