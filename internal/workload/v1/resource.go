// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

// SourceFile represents a golang source code file that contains one or more
// child resource objects.
type SourceFile struct {
	Filename  string
	Children  []ChildResource
	HasStatic bool
}

// ChildResource contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manifests.
type ChildResource struct {
	Name          string
	UniqueName    string
	Group         string
	Version       string
	Kind          string
	StaticContent string
	SourceCode    string
	IncludeCode   string
}

// Resource represents a single input manifest for a given config.
type Resource struct {
	relativeFileName string
	FileName         string `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	Content          []byte `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
}

func (r *Resource) UnmarshalYAML(node *yaml.Node) error {
	r.FileName = node.Value

	return nil
}

func (r *Resource) loadContent() error {
	manifestContent, err := os.ReadFile(r.FileName)
	if err != nil {
		return formatProcessError(r.FileName, err)
	}

	r.Content = manifestContent

	return nil
}

func (r *Resource) extractManifests() []string {
	var manifests []string

	lines := strings.Split(string(r.Content), "\n")

	var manifest string

	for _, line := range lines {
		if strings.TrimRight(line, " ") == "---" {
			if len(manifest) > 0 {
				manifests = append(manifests, manifest)
				manifest = ""
			}
		} else {
			manifest = manifest + "\n" + line
		}
	}

	if len(manifest) > 0 {
		manifests = append(manifests, manifest)
	}

	return manifests
}

func ResourcesFromFiles(resourceFiles []string) []*Resource {
	return getResourcesFromFiles(resourceFiles)
}

func getResourcesFromFiles(resourceFiles []string) []*Resource {
	resources := make([]*Resource, len(resourceFiles))

	for i, resourceFile := range resourceFiles {
		resource := &Resource{
			FileName: resourceFile,
		}

		resources[i] = resource
	}

	return resources
}

func getFuncNames(sourceFiles []SourceFile) (createFuncNames, initFuncNames []string) {
	for _, sourceFile := range sourceFiles {
		for i := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", sourceFile.Children[i].UniqueName)

			if strings.EqualFold(sourceFile.Children[i].Kind, "customresourcedefinition") {
				initFuncNames = append(initFuncNames, funcName)
			}

			createFuncNames = append(createFuncNames, funcName)
		}
	}

	return createFuncNames, initFuncNames
}

func determineSourceFileName(manifestFile string) SourceFile {
	var sourceFile SourceFile
	sourceFile.Filename = filepath.Clean(manifestFile)
	sourceFile.Filename = strings.ReplaceAll(sourceFile.Filename, "/", "_")                              // get filename from path
	sourceFile.Filename = strings.ReplaceAll(sourceFile.Filename, filepath.Ext(sourceFile.Filename), "") // strip ".yaml"
	sourceFile.Filename = strings.ReplaceAll(sourceFile.Filename, ".", "")                               // strip "." e.g. hidden files
	sourceFile.Filename += ".go"                                                                         // add correct file ext
	sourceFile.Filename = utils.ToFileName(sourceFile.Filename)                                          // kebab-case to snake_case

	return sourceFile
}

func expandResources(path string, resources []*Resource) ([]*Resource, error) {
	var expandedResources []*Resource

	for _, r := range resources {
		files, err := Glob(filepath.Join(path, r.FileName))
		if err != nil {
			return []*Resource{}, fmt.Errorf("failed to process glob pattern matching, %w", err)
		}

		for _, f := range files {
			rf, err := filepath.Rel(path, f)
			if err != nil {
				return []*Resource{}, fmt.Errorf("unable to determine relative file path, %w", err)
			}

			res := &Resource{FileName: f, relativeFileName: rf}
			expandedResources = append(expandedResources, res)
		}
	}

	return expandedResources, nil
}

const (
	includeCode = `if %s != %s {
		return []client.Object{}, nil
	}`

	excludeCode = `if %s == %s {
		return []client.Object{}, nil
	}`
)

func (cr *ChildResource) processMarkers(spec *WorkloadSpec) error {
	// obtain the marker results from the input yaml
	_, markerResults, err := inspectMarkersForYAML([]byte(cr.StaticContent), ResourceMarkerType)
	if err != nil {
		return err
	}

	// if we have no resource markers, return
	if len(markerResults) == 0 {
		return nil
	}

	// error if our entire workload spec has no field markers
	if len(spec.CollectionFieldMarkers) == 0 && len(spec.FieldMarkers) == 0 {
		return fmt.Errorf("%w for resource kind %s and name %s",
			ErrResourceMarkerMissingAssociation, cr.Kind, cr.Name)
	}

	var resourceMarker *ResourceMarker

	var fieldMarker interface{}

MARKERS:
	for _, markerResult := range markerResults {
		switch marker := markerResult.Object.(type) {
		case ResourceMarker:
			// associate relevant field markers with this marker
			for _, fm := range spec.FieldMarkers {
				if fm.Name == *marker.Field {
					resourceMarker = &marker
					fieldMarker = fm

					break MARKERS
				}
			}

			// associate relevant collection field markers with this marker
			for _, cm := range spec.CollectionFieldMarkers {
				if cm.Name == *marker.Field {
					resourceMarker = &marker
					fieldMarker = cm

					break MARKERS
				}
			}
		default:
			continue
		}
	}

	if err := resourceMarker.process(fieldMarker); err != nil {
		return err
	}

	if *resourceMarker.Include {
		cr.IncludeCode = fmt.Sprintf(includeCode, resourceMarker.sourceCodeVar, resourceMarker.sourceCodeValue)
	} else {
		cr.IncludeCode = fmt.Sprintf(excludeCode, resourceMarker.sourceCodeVar, resourceMarker.sourceCodeValue)
	}

	return nil
}
