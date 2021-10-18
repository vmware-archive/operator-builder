package v1

import (
	"bytes"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
)

// SourceCode contains information required to generate source code.
type SourceCode struct {
	SpecFields          []*APISpecField
	SourceFiles         *[]SourceFile
	RBACRules           *RBACRules
	OwnershipRules      *[]OwnershipRule
	collection          bool
	collectionResources bool
}

func (sc *SourceCode) processMarkers(manifestFile string) ([]string, error) {
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

	sc.processMarkerResults(markerResults)

	// If processing manifests for collection resources there is no case
	// where there should be collection markers - they will result in
	// code that won't compile.  We will convert collection markers to
	// field markers for the sake of UX.
	if sc.collection && sc.collectionResources {
		// find & replace collection markers with field markers
		manifestContent = []byte(strings.ReplaceAll(string(manifestContent), "!!var collection", "!!var parent"))
	}

	manifests := extractManifests(manifestContent)

	return manifests, nil
}

func (sc *SourceCode) processMarkerResults(markerResults []*inspect.YAMLResult) {
	specFields := make(map[string]*APISpecField)

	for _, markerResult := range markerResults {
		switch r := markerResult.Object.(type) {
		case FieldMarker:
			if sc.collection && !sc.collectionResources {
				continue
			}

			specField := r.ExtractSpecField()

			specFields[r.Name] = specField
		case CollectionFieldMarker:
			if !sc.collection {
				continue
			}

			specField := r.ExtractSpecField()

			specFields[r.Name] = specField
		default:
			continue
		}
	}

	for _, v := range specFields {
		sc.SpecFields = append(sc.SpecFields, v)
	}
}
