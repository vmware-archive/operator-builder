// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
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

// SupportedMarkerDataTypes returns the supported data types that can be used in
// workload markers.
func SupportedMarkerDataTypes() []string {
	return []string{"bool", "string", "int", "int32", "int64", "float32", "float64"}
}

func formatProcessError(manifestFile string, err error) error {
	return fmt.Errorf("error processing file %s; %w", manifestFile, err)
}

//nolint:funlen,gocognit,gocyclo //this will be refactored later
func processMarkers(
	workloadPath string,
	resources []string,
	collection bool,
	collectionResources bool,
) (*SourceCodeTemplateData, error) {
	const dataTypeString = "string"

	results := &SourceCodeTemplateData{
		SourceFiles:    new([]SourceFile),
		RBACRules:      new([]RBACRule),
		OwnershipRules: new([]OwnershipRule),
	}

	specFields := make(map[string]*APISpecField)

	for _, manifestFile := range resources {
		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return nil, formatProcessError(manifestFile, err)
		}

		insp, err := InitializeMarkerInspector()
		if err != nil {
			return nil, formatProcessError(manifestFile, err)
		}

		nodes, markerResults, err := insp.InspectYAML(manifestContent, TransformYAML)
		if err != nil {
			return nil, formatProcessError(manifestFile, err)
		}

		buf := bytes.Buffer{}

		for _, node := range nodes {
			m, err := yaml.Marshal(node)
			if err != nil {
				return nil, formatProcessError(manifestFile, err)
			}

			buf.WriteString("---\n")
			buf.Write(m)
		}

		manifestContent = buf.Bytes()

		for _, markerResult := range markerResults {
			switch r := markerResult.Object.(type) {
			case FieldMarker:
				if collection && !collectionResources {
					continue
				}

				specField := &APISpecField{
					FieldName:         strings.ToTitle(r.Name),
					ManifestFieldName: r.Name,
					DataType:          r.Type.String(),
					APISpecContent: fmt.Sprintf(
						"%s %s `json:\"%s\"`",
						strings.Title(r.Name),
						r.Type,
						r.Name,
					),
				}

				if r.Description != nil {
					specField.DocumentationLines = strings.Split(*r.Description, "\n")
				}

				zv, err := zeroValue(r.Type.String())
				if err != nil {
					return nil, formatProcessError(manifestFile, err)
				}

				specField.ZeroVal = zv

				//nolint:nestif //this will be refactored later
				if r.Default != nil {
					if specField.DataType == dataTypeString {
						specField.DefaultVal = fmt.Sprintf("%q", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.Default)
					} else {
						specField.DefaultVal = fmt.Sprintf("%v", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.Default)
					}
				} else {
					if specField.DataType == dataTypeString {
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.originalValue)
					} else {
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.originalValue)
					}
				}

				specFields[r.Name] = specField
			case CollectionFieldMarker:
				if !collection {
					continue
				}

				specField := &APISpecField{
					FieldName:         strings.ToTitle(r.Name),
					ManifestFieldName: r.Name,
					DataType:          r.Type.String(),
					APISpecContent: fmt.Sprintf(
						"%s %s `json:\"%s\"`",
						strings.Title(r.Name),
						r.Type,
						r.Name,
					),
				}

				if r.Description != nil {
					specField.DocumentationLines = strings.Split(*r.Description, "\n")
				}

				zv, err := zeroValue(r.Type.String())
				if err != nil {
					return nil, formatProcessError(manifestFile, err)
				}

				specField.ZeroVal = zv

				//nolint:nestif //this will be refactored later
				if r.Default != nil {
					if specField.DataType == dataTypeString {
						specField.DefaultVal = fmt.Sprintf("%q", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.Default)
					} else {
						specField.DefaultVal = fmt.Sprintf("%v", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.Default)
					}
				} else {
					if specField.DataType == dataTypeString {
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.originalValue)
					} else {
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.originalValue)
					}
				}

				specFields[r.Name] = specField
			default:
				continue
			}
		}

		if collection && !collectionResources {
			continue
		}

		// determine sourceFile filename
		var sourceFile SourceFile
		sourceFile.Filename = filepath.Base(manifestFile)                // get filename from path
		sourceFile.Filename = strings.Split(sourceFile.Filename, ".")[0] // strip ".yaml"
		sourceFile.Filename += ".go"                                     // add correct file ext
		sourceFile.Filename = utils.ToFileName(sourceFile.Filename)      // kebab-case to snake_case

		var childResources []ChildResource

		manifests := extractManifests(manifestContent)

		for _, manifest := range manifests {
			// If processing manifests for collection resources there is no case
			// where there should be collection markers - they will result in
			// code that won't compile.  We will convert collection markers to
			// field markers for the sake of UX.
			if collection && collectionResources {
				// find & replace collection markers with field markers
				manifest = strings.Replace(
					manifest,
					"!!var collection",
					"!!var parent",
					-1,
				)
			}

			// decode manifest into unstructured data type
			var manifestObject unstructured.Unstructured

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			err := runtime.DecodeInto(decoder, []byte(manifest), &manifestObject)
			if err != nil {
				return nil, formatProcessError(manifestFile, err)
			}

			// generate a unique name for the resource using the kind and name
			resourceUniqueName := strings.Replace(strings.Title(manifestObject.GetName()), "-", "", -1)
			resourceUniqueName = strings.Replace(resourceUniqueName, ".", "", -1)
			resourceUniqueName = strings.Replace(resourceUniqueName, ":", "", -1)
			resourceUniqueName = fmt.Sprintf("%s%s", manifestObject.GetKind(), resourceUniqueName)

			// determine resource group and version
			resourceVersion, resourceGroup := versionGroupFromAPIVersion(manifestObject.GetAPIVersion())

			// determine group and resource for RBAC rule generation
			rbacRulesForManifest(manifestObject.GetKind(), resourceGroup, manifestObject.Object, results.RBACRules)

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

	for _, v := range specFields {
		results.SpecFields = append(results.SpecFields, v)
	}

	// ensure no duplicate file names exist within the source files
	deduplicateFileNames(results)

	return results, nil
}

// deduplicateFileNames dedeplicates the names of the files.  This is because
// we cannot guarantee that files exist in different directories and may have
// naming collisions.
func deduplicateFileNames(templateData *SourceCodeTemplateData) {
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

// zeroValue returns the zero value for the data type as a string.
// It is returned as a string to be used in a template for Go source code.
func zeroValue(val interface{}) (string, error) {
	switch val {
	case "bool":
		return "false", nil
	case "string":
		return "\"\"", nil
	case "int", "int32", "int64", "float32", "float64":
		return "0", nil
	default:
		return "", fmt.Errorf("%w; supported data types: %v", ErrUnsupportedDataType, SupportedMarkerDataTypes())
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

type FieldType int

const (
	FieldUnknownType FieldType = iota
	FieldString
	FieldInt
	FieldBool
)

func (f *FieldType) UnmarshalMarkerArg(in string) error {
	types := map[string]FieldType{
		"":       FieldUnknownType,
		"string": FieldString,
		"int":    FieldInt,
		"bool":   FieldBool,
	}

	if t, ok := types[in]; ok {
		if t == FieldUnknownType {
			return fmt.Errorf("%w, %s into FieldType", ErrUnableToParseFieldType, in)
		}

		*f = t

		return nil
	}

	return fmt.Errorf("%w, %s into FieldType", ErrUnableToParseFieldType, in)
}

func (f FieldType) String() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
	}

	return types[f]
}

type FieldMarker struct {
	Name          string
	Type          FieldType
	Description   *string
	Default       interface{} `marker:",optional"`
	Replace       *string
	originalValue interface{}
}

type CollectionFieldMarker FieldMarker

func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
}
