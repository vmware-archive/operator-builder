// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
)

var ErrProcessManifest = errors.New("error processing manifest file")

// Manifest represents a single input manifest for a given config.
type Manifest struct {
	RelativeFileName string

	Filename string `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	Content  []byte `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
}

// ProcessManifestError is a helper method which returns a consistent format for an
// error when processing a particular manifest file.
func ProcessManifestError(manifest *Manifest, err error) error {
	return fmt.Errorf("%w; %s %s", err, ErrProcessManifest, manifest.Filename)
}

// UnmarshalYAML unmarshals the resources field of a workload configuration.
func (manifest *Manifest) UnmarshalYAML(node *yaml.Node) error {
	manifest.Filename = node.Value

	return nil
}

// LoadContent sets the Content field of the manifest in raw format as []byte.
func (manifest *Manifest) LoadContent(isCollection bool) error {
	manifestContent, err := os.ReadFile(manifest.Filename)
	if err != nil {
		return ProcessManifestError(manifest, err)
	}

	if isCollection {
		// replace all instances of collection markers and collection field markers with regular field markers
		// as a collection marker on a collection is simply a field marker to itself
		content := strings.ReplaceAll(string(manifestContent), markers.CollectionFieldMarkerPrefix, markers.FieldMarkerPrefix)
		content = strings.ReplaceAll(content, markers.ResourceMarkerCollectionFieldName, markers.ResourceMarkerFieldName)

		manifest.Content = []byte(content)
	} else {
		manifest.Content = manifestContent
	}

	return nil
}

// ExpandManifests expands manifests from its globbed pattern and return the resultant manifest
// filenames from the glob.
func ExpandManifests(path string, manifests []*Manifest) ([]*Manifest, error) {
	var expanded []*Manifest

	for i := range manifests {
		files, err := utils.Glob(filepath.Join(path, manifests[i].Filename))
		if err != nil {
			return []*Manifest{}, fmt.Errorf("failed to process glob pattern matching, %w", err)
		}

		for f := range files {
			rf, err := filepath.Rel(path, files[f])
			if err != nil {
				return []*Manifest{}, fmt.Errorf("unable to determine relative file path, %w", err)
			}

			manifest := &Manifest{Filename: files[f], RelativeFileName: rf}
			expanded = append(expanded, manifest)
		}
	}

	return expanded, nil
}

// ExtractManifests extracts the manifests as YAML strings from a manifest with
// existing manifest content.
func (manifest *Manifest) ExtractManifests() []string {
	var manifests []string

	lines := strings.Split(string(manifest.Content), "\n")

	var content string

	for _, line := range lines {
		if strings.TrimRight(line, " ") == "---" {
			if len(content) > 0 {
				manifests = append(manifests, content)
				content = ""
			}
		} else {
			content = content + "\n" + line
		}
	}

	if len(content) > 0 {
		manifests = append(manifests, content)
	}

	return manifests
}

// GetManifests returns the manifest objects from the given manifest file paths
// as string inputs.
func GetManifests(manifestFiles []string) []*Manifest {
	manifests := make([]*Manifest, len(manifestFiles))

	for i, manifestFile := range manifestFiles {
		manifest := &Manifest{
			Filename: manifestFile,
		}

		manifests[i] = manifest
	}

	return manifests
}
