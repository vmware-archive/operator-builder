// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
)

const (
	FieldMarkerType MarkerType = iota
	CollectionMarkerType
	ResourceMarkerType
)

type MarkerType int

type FieldMarker struct {
	Name          string
	Type          FieldType
	Description   *string
	Default       interface{} `marker:",optional"`
	Replace       *string
	originalValue interface{}
}

// TODO: pointer to string/bool
type ResourceMarker struct {
	Field           string `marker:",optional"`
	CollectionField string `marker:",optional"`
	Value           interface{}
	Exclude         bool `marker:",optional"`
	Include         bool `marker:",optional"`

	sourceCodeVar   string
	sourceCodeValue string
}

var ErrMismatchedMarkerTypes = errors.New("resource marker and field marker have mismatched types")
var ErrResourceMarkerUnknownValueType = errors.New("resource marker 'value' is of unknown type")
var ErrResourceMarkerHasIncludeExclude = errors.New("resource marker cannot have both 'include' and 'exclude' specified")
var ErrResourceMarkerMissingAssociation = errors.New("resource marker cannot find an associated field marker")
var ErrResourceMarkerMissingFieldValue = errors.New("resource marker missing 'field' or 'value'")
var ErrFieldMarkerInvalidType = errors.New("field marker type is invalid")

func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
}

type CollectionFieldMarker FieldMarker

func (cfm CollectionFieldMarker) String() string {
	return fmt.Sprintf("CollectionFieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		cfm.Name,
		cfm.Type,
		*cfm.Description,
		cfm.Default,
	)
}

func (rm ResourceMarker) String() string {
	return fmt.Sprintf("ResourceMarker{Field: %s CollectionField: %s Value: %v Exclude: %v Include: %v}",
		rm.Field,
		rm.CollectionField,
		rm.Value,
		rm.Exclude,
		rm.Include,
	)
}

func InitializeMarkerInspector(markerTypes ...MarkerType) (*inspect.Inspector, error) {
	registry := marker.NewRegistry()

	fieldMarker, err := marker.Define("+operator-builder:field", FieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	collectionMarker, err := marker.Define("+operator-builder:collection:field", CollectionFieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	resourceMarker, err := marker.Define("+operator-builder:resource", ResourceMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	for _, markerType := range markerTypes {
		switch markerType {
		case FieldMarkerType:
			registry.Add(fieldMarker)
		case CollectionMarkerType:
			registry.Add(collectionMarker)
		case ResourceMarkerType:
			registry.Add(resourceMarker)
		}
	}

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

		replaceText := strings.TrimSuffix(r.MarkerText, "\n")
		replaceText = strings.ReplaceAll(replaceText, "\n", "\n#")

		key.FootComment = ""

		switch t := r.Object.(type) {
		case FieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = key.HeadComment + "\n# " + *t.Description
			}

			key.HeadComment = strings.ReplaceAll(key.HeadComment, replaceText, "controlled by field: "+t.Name)
			value.LineComment = strings.ReplaceAll(value.LineComment, replaceText, "controlled by field: "+t.Name)

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
				key.HeadComment = "# " + *t.Description
			}

			key.HeadComment = strings.ReplaceAll(key.HeadComment, replaceText, "controlled by collection field: "+t.Name)
			value.LineComment = strings.ReplaceAll(value.LineComment, replaceText, "controlled by collection field: "+t.Name)

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

func containsMarkerType(s []MarkerType, e MarkerType) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func inspectMarkersForYAML(yamlContent []byte, markerTypes ...MarkerType) ([]*yaml.Node, []*inspect.YAMLResult, error) {
	insp, err := InitializeMarkerInspector(markerTypes...)
	if err != nil {
		return nil, nil, err
	}

	return insp.InspectYAML(yamlContent, TransformYAML)
}

func (rm *ResourceMarker) setSourceCodeVar(prefix string) {
	if rm.Field != "" {
		rm.sourceCodeVar = fmt.Sprintf("%s.%s", prefix, strings.Title(rm.Field))
	} else {
		rm.sourceCodeVar = fmt.Sprintf("%s.%s", prefix, strings.Title(rm.CollectionField))
	}
}

func (rm *ResourceMarker) hasField() bool {
	return rm.Field != "" || rm.CollectionField != ""
}

func (rm *ResourceMarker) hasValue() bool {
	return rm.Value != ""
}

func (rm *ResourceMarker) process(fieldMarker interface{}) error {
	var fieldType string

	// ensure that a field and value exist
	if !rm.hasField() || !rm.hasValue() {
		return ErrResourceMarkerMissingFieldValue
	}

	// determine if our associated field marker is a collection or regular field marker and
	// set appropriate variables
	switch marker := fieldMarker.(type) {
	case *CollectionFieldMarker:
		fieldType = marker.Type.String()

		rm.setSourceCodeVar("collection.Spec")
	case *FieldMarker:
		fieldType = marker.Type.String()

		rm.setSourceCodeVar("parent.Spec")
	default:
		return fmt.Errorf("%w; %T", ErrFieldMarkerInvalidType, fieldMarker)
	}

	// validate that the field types match
	switch value := rm.Value.(type) {
	case string:
		if fieldType != "string" {
			return fmt.Errorf("%w; expected: string, got: %s", ErrMismatchedMarkerTypes, fieldType)
		}

		rm.sourceCodeValue = fmt.Sprintf("%q", value)
	case int:
		if fieldType != "int" {
			return fmt.Errorf("%w; expected: int, got: %s", ErrMismatchedMarkerTypes, fieldType)
		}

		rm.sourceCodeValue = fmt.Sprintf("%v", value)
	case bool:
		if fieldType != "bool" {
			return fmt.Errorf("%w; expected: bool, got: %s", ErrMismatchedMarkerTypes, fieldType)
		}

		rm.sourceCodeValue = fmt.Sprintf("%v", value)
	default:
		return ErrResourceMarkerUnknownValueType
	}

	return nil
}
