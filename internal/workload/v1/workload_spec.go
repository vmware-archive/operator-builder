// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
)

// WorkloadSpec contains information required to generate source code.
type WorkloadSpec struct {
	Resources           []string `json:"resources" yaml:"resources"`
	APISpecFields       []*APISpecField
	SourceFiles         *[]SourceFile
	RBACRules           *RBACRules
	OwnershipRules      *OwnershipRules
	collection          bool
	collectionResources bool
}

func (ws *WorkloadSpec) Init() {
	ws.APISpecFields = []*APISpecField{}

	ws.OwnershipRules = &OwnershipRules{}
	ws.RBACRules = &RBACRules{}
	ws.SourceFiles = &[]SourceFile{}
}

func (ws *WorkloadSpec) processManifests(workloadPath string, collection, collectionResources bool) error {
	ws.Init()

	specFields := make(map[string]*APISpecField)

	for _, manifestFile := range ws.Resources {
		// capture entire resource manifest file content
		manifests, err := ws.processMarkers(filepath.Join(filepath.Dir(workloadPath), manifestFile), specFields)
		if err != nil {
			return err
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
				return formatProcessError(manifestFile, err)
			}

			// generate a unique name for the resource using the kind and name
			resourceUniqueName := generateUniqueResourceName(manifestObject)
			// determine resource group and version
			resourceVersion, resourceGroup := versionGroupFromAPIVersion(manifestObject.GetAPIVersion())

			// determine group and resource for RBAC rule generation
			ws.RBACRules.addRulesForManifest(manifestObject.GetKind(), resourceGroup, manifestObject.Object)

			ws.OwnershipRules.addOrUpdateOwnership(
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
				return formatProcessError(manifestFile, err)
			}

			// add the source code to the resource
			resource.SourceCode = resourceDefinition
			resource.StaticContent = manifest

			childResources = append(childResources, resource)
		}

		sourceFile.Children = childResources

		if ws.SourceFiles == nil {
			ws.SourceFiles = &[]SourceFile{}
		}

		*ws.SourceFiles = append(*ws.SourceFiles, sourceFile)
	}

	for _, v := range specFields {
		ws.APISpecFields = append(ws.APISpecFields, v)
	}

	// ensure no duplicate file names exist within the source files
	ws.deduplicateFileNames()

	return nil
}

func (ws *WorkloadSpec) processMarkers(manifestFile string, specFields map[string]*APISpecField) ([]string, error) {
	// capture entire resource manifest file content
	manifestContent, err := ioutil.ReadFile(manifestFile)
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

	ws.processMarkerResults(markerResults, specFields)

	// If processing manifests for collection resources there is no case
	// where there should be collection markers - they will result in
	// code that won't compile.  We will convert collection markers to
	// field markers for the sake of UX.
	if ws.collection && ws.collectionResources {
		// find & replace collection markers with field markers
		manifestContent = []byte(strings.ReplaceAll(string(manifestContent), "!!var collection", "!!var parent"))
	}

	manifests := extractManifests(manifestContent)

	return manifests, nil
}

func (ws *WorkloadSpec) processMarkerResults(markerResults []*inspect.YAMLResult, specFields map[string]*APISpecField) {
	for _, markerResult := range markerResults {
		switch r := markerResult.Object.(type) {
		case FieldMarker:
			if ws.collection && !ws.collectionResources {
				continue
			}

			specField := r.ExtractSpecField()

			specFields[r.Name] = specField
		case CollectionFieldMarker:
			if !ws.collection {
				continue
			}

			specField := r.ExtractSpecField()

			specFields[r.Name] = specField
		default:
			continue
		}
	}
}

// deduplicateFileNames dedeplicates the names of the files.  This is because
// we cannot guarantee that files exist in different directories and may have
// naming collisions.
func (ws *WorkloadSpec) deduplicateFileNames() {
	// create a slice to track existing fileNames and preallocate an existing
	// known conflict
	fileNames := make([]string, len(*ws.SourceFiles)+1)
	fileNames[len(fileNames)-1] = "resources.go"

	// dereference the sourcefiles
	sourceFiles := *ws.SourceFiles

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
