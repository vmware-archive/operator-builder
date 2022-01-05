// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package api

import (
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var (
	_ machinery.Template = &Kind{}
	_ machinery.Template = &KindLatest{}
	_ machinery.Inserter = &KindUpdater{}
)

// Code Markers (associated with below fragments).
const (
	kindImportsMarker       = "operator-builder:imports"
	kindGroupVersionsMarker = "operator-builder:groupversions"
)

// Code Fragments (associated with above markers).
const (
	kindImportsFragment = `%s "%s/apis/%s/%s"
`
	kindGroupVersionsFragment = `%s.GroupVersion,
`
)

// Kind scaffolds the file that defines specific information related to kind regardless of
// the API version.
type Kind struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
}

func (f *Kind) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		fmt.Sprintf("%s.go", strings.ToLower(f.Resource.Kind)),
	)

	f.TemplateBody = fmt.Sprintf(
		kindTemplate,
		machinery.NewMarkerFor(f.Path, kindImportsMarker),
		machinery.NewMarkerFor(f.Path, kindGroupVersionsMarker),
	)

	return nil
}

// KindLatest scaffolds the file that defines specific information related to the latest API version
// of a specific kind.
type KindLatest struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.RepositoryMixin

	PackageName string
}

func (f *KindLatest) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		fmt.Sprintf("%s_latest.go", strings.ToLower(f.Resource.Kind)),
	)

	f.TemplateBody = kindLatestTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

// KindUpdater updates the file with any new version information.
type KindUpdater struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.RepositoryMixin
}

// GetPath implements file.Builder interface.
func (f *KindUpdater) GetPath() string {
	return filepath.Join(
		"apis",
		f.Resource.Group,
		fmt.Sprintf("%s.go", strings.ToLower(f.Resource.Kind)),
	)
}

// GetIfExistsAction implements file.Builder interface.
func (*KindUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

// GetMarkers implements file.Inserter interface.
func (f *KindUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), kindImportsMarker),
		machinery.NewMarkerFor(f.GetPath(), kindGroupVersionsMarker),
	}
}

// GetCodeFragments implements file.Inserter interface.
func (f *KindUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	versionGroup := fmt.Sprintf("%s%s", f.Resource.Version, f.Resource.Group)

	// Generate imports code fragments
	imports := make([]string, 0)
	imports = append(imports, fmt.Sprintf(kindImportsFragment,
		versionGroup,
		f.Repo,
		f.Resource.Group,
		f.Resource.Version,
	))

	// Generate groupVersions code fragments
	groupVersions := make([]string, 0)
	groupVersions = append(groupVersions, fmt.Sprintf(kindGroupVersionsFragment, versionGroup))

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), kindImportsMarker)] = imports
	}

	if len(groupVersions) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), kindGroupVersionsMarker)] = groupVersions
	}

	return fragments
}

const (
	kindTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	%s
	
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// {{ .Resource.Kind }}GroupVersions returns all group version objects associated with this kind.
func {{ .Resource.Kind }}GroupVersions() []schema.GroupVersion {
	return []schema.GroupVersion{
		%s
	}
}
`
	kindLatestTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	{{ .Resource.Version }}{{ .Resource.Group }} "{{ .Repo }}/apis/{{ .Resource.Group }}/{{ .Resource.Version }}"
	{{ .Resource.Version }}{{ lower .Resource.Kind }} "{{ .Resource.Path }}/{{ .PackageName }}"
)

// Code generated by operator-builder. DO NOT EDIT.

// {{ .Resource.Kind }}LatestGroupVersion returns the latest group version object associated with this
// particular kind.
var {{ .Resource.Kind }}LatestGroupVersion = {{ .Resource.Version }}{{ .Resource.Group }}.GroupVersion

// {{ .Resource.Kind }}LatestSample returns the latest sample manifest associated with this
// particular kind.
var {{ .Resource.Kind }}LatestSample = {{ .Resource.Version }}{{ lower .Resource.Kind }}.Sample(false)
`
)
