package v1

import (
	"bytes"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

// SourceCode is a collection of variables used to generate source code.
type SourceCode struct {
	SpecFields     []*APISpecField
	SourceFiles    *[]SourceFile
	RBACRules      *RBACRules
	OwnershipRules *[]OwnershipRule
}

func (results *SourceCode) processMarkers(manifestFile string, collection, collectionResources bool) ([]string, error) {
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

	specFields := processMarkerResults(markerResults, collection, collectionResources)

	for _, v := range specFields {
		results.SpecFields = append(results.SpecFields, v)
	}

	// If processing manifests for collection resources there is no case
	// where there should be collection markers - they will result in
	// code that won't compile.  We will convert collection markers to
	// field markers for the sake of UX.
	if collection && collectionResources {
		// find & replace collection markers with field markers
		manifestContent = []byte(strings.ReplaceAll(string(manifestContent), "!!var collection", "!!var parent"))
	}

	manifests := extractManifests(manifestContent)

	return manifests, nil
}
