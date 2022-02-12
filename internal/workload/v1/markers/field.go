// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"errors"
	"fmt"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
)

var (
	ErrFieldMarkerReserved    = errors.New("field marker cannot be used and is reserved for internal purposes")
	ErrFieldMarkerInvalidType = errors.New("field marker type is invalid")
)

const (
	FieldMarkerPrefix = "+operator-builder:field"
	FieldSpecPrefix   = "parent.Spec"
)

// FieldMarker is an object which represents a marker that is associated with a field
// that exsists within a manifest.  A FieldMarker is discovered when a manifest is parsed
// and matches the constants defined by the fieldMarker constant above.
type FieldMarker struct {
	// inputs from the marker itself
	Name        string
	Type        FieldType
	Description *string
	Default     interface{} `marker:",optional"`
	Replace     *string

	// other values which we use to pass information
	ForCollection bool
	OriginalValue interface{}
}

//nolint:gocritic //needed to implement string interface
func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
}

// defineFieldMarker will define a FieldMarker and add it a registry of markers.
func defineFieldMarker(registry *marker.Registry) error {
	fieldMarker, err := marker.Define(FieldMarkerPrefix, FieldMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(fieldMarker)

	return nil
}

//
// FieldMarker Processor interface methods.
//
func (fm *FieldMarker) GetName() string {
	return fm.Name
}

func (fm *FieldMarker) GetDescription() string {
	return *fm.Description
}

func (fm *FieldMarker) GetFieldType() FieldType {
	return fm.Type
}

func (fm *FieldMarker) GetReplaceText() string {
	return *fm.Replace
}

func (fm *FieldMarker) GetSpecPrefix() string {
	return CollectionFieldSpecPrefix
}

func (fm *FieldMarker) SetOriginalValue(value string) {
	fm.OriginalValue = &value
}

func (fm *FieldMarker) SetDescription(description string) {
	fm.Description = &description
}
