// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

var (
	ErrUnsupportedDataType    = errors.New("unsupported data type in workload marker")
	ErrUnableToParseFieldType = errors.New("unable to parse field")
)

func formatProcessError(manifestFile string, err error) error {
	return fmt.Errorf("error processing file %s; %w", manifestFile, err)
}

func processManifests(
	workloadPath string,
	resources []string,
	collection bool,
	collectionResources bool,
) (*SourceCode, error) {
	results := &SourceCode{
		SourceFiles:    new([]SourceFile),
		RBACRules:      new(RBACRules),
		OwnershipRules: new([]OwnershipRule),
	}

	for _, manifestFile := range resources {
		// capture entire resource manifest file content
		manifests, err := results.processMarkers(filepath.Join(filepath.Dir(workloadPath), manifestFile), collection, collectionResources)
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

			// determine group and kind for ownership rule generation
			newOwnershipRule := OwnershipRule{
				Version: manifestObject.GetAPIVersion(),
				Kind:    manifestObject.GetKind(),
				CoreAPI: isCoreAPI(resourceGroup),
			}

			ownershipExists := versionKindRecorded(results.OwnershipRules, &newOwnershipRule)
			if !ownershipExists {
				*results.OwnershipRules = append(*results.OwnershipRules, newOwnershipRule)
			}

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

func InitializeMarkerInspector() (*inspect.Inspector, error) {
	registry := marker.NewRegistry()

	fieldMarker, err := marker.Define("+operator-builder:field", FieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	collectionMarker, err := marker.Define("+operator-builder:collection:field", CollectionFieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	registry.Add(fieldMarker)
	registry.Add(collectionMarker)

	return inspect.NewInspector(registry), nil
}

func TransformYAML(results ...*inspect.YAMLResult) error {
	const varTag = "!!var"

	const strTag = "!!str"

	var key *yaml.Node

	var value *yaml.Node

	for _, r := range results {
		if len(r.Nodes) > 1 {
			key = r.Nodes[0]
			value = r.Nodes[1]
		} else {
			key = r.Nodes[0]
			value = r.Nodes[0]
		}

		key.HeadComment = ""
		key.FootComment = ""
		value.LineComment = ""

		switch t := r.Object.(type) {
		case FieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

			t.originalValue = value.Value

			if t.Replace != nil {
				value.Tag = strTag

				re, err := regexp.Compile(*t.Replace)
				if err != nil {
					return fmt.Errorf("unable to convert %s to regex, %w", *t.Replace, err)
				}

				value.Value = re.ReplaceAllString(value.Value, fmt.Sprintf("!!start parent.Spec.%s !!end", strings.Title((t.Name))))
			} else {
				value.Tag = varTag
				value.Value = fmt.Sprintf("parent.Spec." + strings.Title(t.Name))
			}

			r.Object = t

		case CollectionFieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

			t.originalValue = value.Value

			if t.Replace != nil {
				value.Tag = strTag

				re, err := regexp.Compile(*t.Replace)
				if err != nil {
					return fmt.Errorf("unable to convert %s to regex, %w", *t.Replace, err)
				}

				value.Value = re.ReplaceAllString(value.Value, fmt.Sprintf("!!start collection.Spec.%s !!end", strings.Title((t.Name))))
			} else {
				value.Tag = varTag
				value.Value = fmt.Sprintf("collection.Spec." + strings.Title(t.Name))
			}

			r.Object = t
		}
	}

	return nil
}

type Marker interface {
	GetType() string
	ExtractSpecField()
}

func processMarkerResults(markerResults []*inspect.YAMLResult, collection, collectionResources bool) map[string]*APISpecField {
	specFields := make(map[string]*APISpecField)

	for _, markerResult := range markerResults {
		switch r := markerResult.Object.(type) {
		case FieldMarker:
			if collection && !collectionResources {
				continue
			}

			specField := r.ExtractSpecField()

			specFields[r.Name] = specField
		case CollectionFieldMarker:
			if !collection {
				continue
			}

			specField := r.ExtractSpecField()

			specFields[r.Name] = specField
		default:
			continue
		}
	}

	return specFields
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
