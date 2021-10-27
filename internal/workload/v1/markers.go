// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
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

	for _, markerType := range markerTypes {
		switch markerType {
		case FieldMarkerType:
			registry.Add(fieldMarker)
		case CollectionMarkerType:
			registry.Add(collectionMarker)
		}
	}

	return inspect.NewInspector(registry), nil
}

func TransformYAML(results ...*inspect.YAMLResult) error {
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
			setDescription(key, t.Description, t.Name)

			originalValue, err := getOrginalValue(value)
			if err != nil {
				return fmt.Errorf("unable to get original value, %w", err)
			}

			t.originalValue = originalValue

			if err := insertVariable(value, t.Name, "parent", t.Replace); err != nil {
				return err
			}

			r.Object = t

		case CollectionFieldMarker:
			setDescription(key, t.Description, t.Name)

			originalValue, err := getOrginalValue(value)
			if err != nil {
				return fmt.Errorf("unable to get original value, %w", err)
			}

			t.originalValue = originalValue

			if err := insertVariable(value, t.Name, "collection", t.Replace); err != nil {
				return err
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

func insertVariable(value *yaml.Node, name, root string, replace *string) error {
	const varTag = "!!var"

	if replace != nil {
		re, err := regexp.Compile(*replace)
		if err != nil {
			return fmt.Errorf("unable to convert %s to regex, %w", *replace, err)
		}

		if value.Content != nil {
			for _, node := range value.Content {
				node.Value = re.ReplaceAllString(node.Value, fmt.Sprintf("!!start %s.Spec.%s !!end", root, strings.Title((name))))
			}
		} else {
			value.Value = re.ReplaceAllString(value.Value, fmt.Sprintf("!!start %s.Spec.%s !!end", root, strings.Title((name))))
		}
	} else {
		value.Tag = varTag
		value.Kind = yaml.ScalarNode
		value.Value = fmt.Sprintf("%s.Spec.%s", root, strings.Title(name))
		value.Content = nil
	}

	return nil
}

func getOrginalValue(value *yaml.Node) (interface{}, error) {
	var v interface{}
	if err := value.Decode(&v); err != nil {
		return nil, fmt.Errorf("unable to decode value, %w", err)
	}

	return v, nil
}

func setDescription(key *yaml.Node, description *string, name string) {
	if description != nil {
		*description = strings.TrimPrefix(*description, "\n")
		key.HeadComment = "# " + *description + ", controlled by " + name
	}
}
