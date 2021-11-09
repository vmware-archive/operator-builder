// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package e2e

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

const (
	e2eTestWorkloadPath = "test/e2e/%s_workloads_test.go"
	importMarker        = "operator-builder:imports"
	testWorkloadsMarker = "operator-builder:testworkloads"
)

var (
	_ machinery.Template = &WorkloadTest{}
	_ machinery.Inserter = &WorkloadTestUpdater{}
)

// WorkloadTest adds API-specific scaffolding for each workload test case.
type WorkloadTest struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.DomainMixin
	machinery.RepositoryMixin
	machinery.ComponentConfigMixin
	machinery.ResourceMixin
}

func (f *WorkloadTest) SetTemplateDefaults() error {
	f.Path = fmt.Sprintf(e2eTestWorkloadPath, f.Resource.Version)

	f.TemplateBody = fmt.Sprintf(e2eWorkloadsTemplate,
		machinery.NewMarkerFor(f.Path, importMarker),
		machinery.NewMarkerFor(f.Path, testWorkloadsMarker),
	)

	return nil
}

type WorkloadTestUpdater struct {
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	HasChildResources bool
	IsStandalone      bool
	IsComponent       bool
	IsCollection      bool
	IsClusterScoped   bool
	Collection        *workloadv1.WorkloadCollection
	PackageName       string

	TesterName           string
	TesterNamespace      string
	TesterSamplePath     string
	TesterCollectionName string
}

func (f *WorkloadTestUpdater) GetPath() string {
	return fmt.Sprintf(e2eTestWorkloadPath, f.Resource.Version)
}

func (*WorkloadTestUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

func (f *WorkloadTestUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(e2eTestWorkloadPath, importMarker),
		machinery.NewMarkerFor(e2eTestWorkloadPath, testWorkloadsMarker),
	}
}

func (f *WorkloadTestUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	const options = 3

	fragments := make(machinery.CodeFragmentsMap, options)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate import code fragments
	imports := make([]string, 0)
	imports = append(imports, fmt.Sprintf(importCodeFragment,
		f.Resource.ImportAlias(), f.Resource.Path,
		f.Resource.Path, strings.ToLower(f.Resource.Kind)),
	)

	// Set variables needed for templating
	f.TesterNamespace = getTesterNamespace(f.IsClusterScoped, f.Resource)
	f.TesterSamplePath = getTesterSamplePath(f.Resource)
	f.TesterName = getTesterName(f.Resource)

	if f.Collection != nil {
		f.TesterCollectionName = getTesterCollectionName(f.Collection)
	}

	t, err := template.New("testerTemplate").Parse(e2eWorkloadFragment)
	if err != nil {
		panic(err)
	}

	// working buffer
	workloadTestFragmentParsed := &bytes.Buffer{}

	err = t.Execute(workloadTestFragmentParsed, f)
	if err != nil {
		panic(err)
	}

	// Generate test code fragments
	workloadTestFragments := make([]string, 0)
	workloadTestFragments = append(workloadTestFragments, workloadTestFragmentParsed.String())

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(e2eTestWorkloadPath, importMarker)] = imports
	}

	if len(workloadTestFragments) != 0 {
		fragments[machinery.NewMarkerFor(e2eTestWorkloadPath, testWorkloadsMarker)] = workloadTestFragments
	}

	return fragments
}

const (
	importCodeFragment = `%s "%s"
		"%s/%s"
	`
	e2eWorkloadFragment = `

	//
	// {{ .TesterName }} tests
	//
	func {{ .TesterName }}ChildrenFuncs(tester *E2ETest) error {
		tester.children = make([]metav1.Object, len({{ .PackageName }}.CreateFuncs))

		if len(tester.children) == 0 {
			return nil
		}

		workload, ok := tester.workload.(*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }})
		if !ok {
			return fmt.Errorf("could not convert metav1.Object to {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}")
		}

		{{ if .IsComponent -}}
		collection, ok := tester.collectionTester.workload.(*{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }})
		if !ok {
			return fmt.Errorf("could not convert metav1.Object to {{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }}")
		}
		{{ end }}

		for i, f := range {{ .PackageName }}.CreateFuncs {
			resource, err := f(workload{{ if .IsComponent }}, collection){{ else }}){{ end }}
			if err != nil {
				return fmt.Errorf("unable to create object in memory; %s", err)
			}

			tester.children[i] = resource
		}

		return nil
	}

	var {{ .TesterName }}Test = &E2ETest{
		namespace:          "{{ .TesterNamespace }}",
		unstructured:       &unstructured.Unstructured{},
		workload:           &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{},
		sampleManifestFile: "{{ .TesterSamplePath }}",
		getChildrenFunc:    {{ .TesterName }}ChildrenFuncs,
		{{ if .IsComponent -}}
		collectionTester:   {{ .TesterCollectionName }}Test,     
		{{ end }}
	}

	{{ if .IsCollection -}}
	func (s *E2ECollectionTestSuite) Test_{{ .TesterName }}() {
	{{ else }}
	func (s *E2EComponentTestSuite) Test_{{ .TesterName }}() {
	{{ end -}}
		// setup
		s.suite.tests = append(s.suite.tests, {{ .TesterName }}Test)
		{{ .TesterName }}Test.suite = &s.suite
		require.NoErrorf(s.T(), setupTest({{ .TesterName }}Test), "failed to setup test")

		// create the custom resource
		require.NoErrorf(s.T(), createCustomResource({{ .TesterName }}Test), "failed to create custom resource")

		// double-check that the child resources are ready
		require.NoErrorf(s.T(), waitForChildResources({{ .TesterName }}Test), "failed to wait for custom resource child resources")

		// test the deletion of a child object
		require.NoErrorf(s.T(), testDeleteChildResource({{ .TesterName }}Test), "failed to reconcile deletion of a child resource")

		// test the update of a child object
		// TODO: need immutable fields so that we can predict which managed fields we can modify to test reconciliation
		// see https://github.com/vmware-tanzu-labs/operator-builder/issues/67

		// test the update of a parent object
		// TODO: need immutable fields so that we can predict which managed fields we can modify to test reconciliation
		// see https://github.com/vmware-tanzu-labs/operator-builder/issues/67
	}
	`
)

const e2eWorkloadsTemplate = `// +build e2e_test

{{ .Boilerplate }}

package e2e_test

import (
	"fmt"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	%s
)

%s
`

func getTesterSamplePath(r *resource.Resource) string {
	return strings.Join([]string{
		"../..",
		"config",
		"samples",
		fmt.Sprintf(
			"%s_%s_%s.yaml",
			r.Group,
			r.Version,
			utils.ToFileName(r.Kind),
		)}, "/",
	)
}

func getTesterNamespace(clusterScoped bool, r *resource.Resource) (namespace string) {
	if !clusterScoped {
		namespaceElements := []string{
			"test",
			strings.ToLower(r.Group),
			strings.ToLower(r.Version),
			strings.ToLower(r.Kind),
		}
		namespace = strings.Join(namespaceElements, "-")
	}

	return namespace
}

func getTesterName(r *resource.Resource) string {
	return r.ImportAlias() + r.Kind
}

func getTesterCollectionName(collection *workloadv1.WorkloadCollection) string {
	return strings.Join([]string{
		strings.ToLower(collection.Spec.API.Group),
		strings.ToLower(collection.GetAPIVersion()),
		collection.Spec.API.Kind,
	}, "")
}
