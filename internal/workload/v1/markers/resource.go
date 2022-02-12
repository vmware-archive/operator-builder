// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"errors"
	"fmt"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
)

var (
	ErrResourceMarkerInvalid            = errors.New("resource marker is invalid")
	ErrResourceMarkerCount              = errors.New("expected only 1 resource marker")
	ErrResourceMarkerAssociation        = errors.New("unable to associate resource marker with field marker")
	ErrResourceMarkerTypeMismatch       = errors.New("resource marker and field marker have mismatched types")
	ErrResourceMarkerInvalidType        = errors.New("expected resource marker type")
	ErrResourceMarkerUnknownValueType   = errors.New("resource marker 'value' is of unknown type")
	ErrResourceMarkerMissingFieldValue  = errors.New("resource marker missing 'collectionField', 'field' or 'value'")
	ErrResourceMarkerMissingInclude     = errors.New("resource marker missing 'include' value")
	ErrResourceMarkerMissingFieldMarker = errors.New("resource marker has no associated 'field' or 'collectionField' marker")
)

const (
	ResourceMarkerPrefix              = "+operator-builder:resource"
	ResourceMarkerCollectionFieldName = "collectionField"
	ResourceMarkerFieldName           = "field"
)

// If we have a valid resource marker,  we will either include or exclude the
// related object based on the inputs on the resource marker itself.  These are
// the resultant code snippets based on that logic.
const (
	includeCode = `if %s != %s {
		return []client.Object{}, nil
	}`

	excludeCode = `if %s == %s {
		return []client.Object{}, nil
	}`
)

// ResourceMarker is an object which represents a marker for an entire resource.  It
// allows actions against a resource.  A ResourceMarker is discovered when a manifest
// is parsed and matches the constants defined by the collectionFieldMarker
// constant above.
type ResourceMarker struct {
	// inputs from the marker itself
	Field           *string
	CollectionField *string
	Value           interface{}
	Include         *bool

	// other field which we use to pass information
	IncludeCode     string
	sourceCodeVar   string
	sourceCodeValue string
	fieldMarker     FieldMarkerProcessor
}

//nolint:gocritic //needed to implement string interface
func (rm ResourceMarker) String() string {
	return fmt.Sprintf("ResourceMarker{Field: %s CollectionField: %s Value: %v Include: %v}",
		*rm.Field,
		*rm.CollectionField,
		rm.Value,
		*rm.Include,
	)
}

// defineResourceMarker will define a ResourceMarker and add it a registry of markers.
func defineResourceMarker(registry *marker.Registry) error {
	resourceMarker, err := marker.Define(ResourceMarkerPrefix, ResourceMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(resourceMarker)

	return nil
}

// Process will process a resource marker from a collection of collection field markers
// and field markers, associate them together and set the appropriate fields.
func (rm *ResourceMarker) Process(markers *MarkerCollection) error {
	// associate field markers from a collection of markers to this resource marker
	rm.associateFieldMarker(markers)

	if err := rm.validate(); err != nil {
		return fmt.Errorf("%w; %s", err, ErrResourceMarkerInvalid)
	}

	// set the source code and return
	if err := rm.setSourceCode(); err != nil {
		return fmt.Errorf("%w; error setting source code for resource marker: %v", err, rm)
	}

	return nil
}

// hasField determines whether or not a parsed resource marker has either a field
// or a collection field.  One or the other is needed for processing a resource
// marker.
func (rm *ResourceMarker) hasField() bool {
	var hasField, hasCollectionField bool

	if rm.Field != nil {
		if *rm.Field != "" {
			hasField = true
		}
	}

	if rm.CollectionField != nil {
		if *rm.CollectionField != "" {
			hasCollectionField = true
		}
	}

	return hasField || hasCollectionField
}

// hasValue determines whether or not a parsed resource marker has a value
// to check against.
func (rm *ResourceMarker) hasValue() bool {
	return rm.Value != nil
}

// associateFieldMarker will associate a resource marker with one of a field
// marker or collection marker.
func (rm *ResourceMarker) associateFieldMarker(markers *MarkerCollection) {
	// return immediately if the marker collection we are trying to associate is empty
	if len(markers.CollectionFieldMarkers) == 0 && len(markers.FieldMarkers) == 0 {
		return
	}

	// associate first relevant field marker with this marker
	for _, fm := range markers.FieldMarkers {
		if rm.Field != nil {
			if fm.Name == *rm.Field {
				rm.fieldMarker = fm

				return
			}
		}

		if fm.ForCollection {
			if rm.CollectionField != nil {
				if fm.Name == *rm.CollectionField {
					rm.fieldMarker = fm

					return
				}
			}
		}
	}

	// associate first relevant collection field marker with this marker
	for _, cm := range markers.CollectionFieldMarkers {
		if rm.CollectionField != nil {
			if cm.Name == *rm.CollectionField {
				rm.fieldMarker = cm

				return
			}
		}
	}
}

// validate checks for a valid resource marker and returns an error if the
// resource marker is invalid.
func (rm *ResourceMarker) validate() error {
	// check include field for a provided value
	// NOTE: this field is mandatory now, but could be optional later, so we return
	// an error here rather than using a pointer to a bool to control the mandate.
	if rm.Include == nil {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingInclude, rm)
	}

	if rm.fieldMarker == nil {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingFieldMarker, rm)
	}

	// ensure that both a field and value exist
	if !rm.hasField() || !rm.hasValue() {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingFieldValue, rm)
	}

	return nil
}

func (rm *ResourceMarker) setSourceCode() error {
	// set the source code variable
	rm.sourceCodeVar = getSourceCodeVariable(rm.fieldMarker)

	// set the source code value and ensure the types match
	switch value := rm.Value.(type) {
	case string, int, bool:
		fieldMarkerType := rm.fieldMarker.GetFieldType().String()
		resourceMarkerType := fmt.Sprintf("%T", value)

		if fieldMarkerType != resourceMarkerType {
			return fmt.Errorf("%w; expected: %s, got: %s for marker %s",
				ErrResourceMarkerTypeMismatch,
				resourceMarkerType,
				fieldMarkerType,
				rm,
			)
		}

		rm.sourceCodeValue = fmt.Sprintf("%q", value)
	default:
		return ErrResourceMarkerUnknownValueType
	}

	// set the include code for this marker
	if *rm.Include {
		rm.IncludeCode = fmt.Sprintf(includeCode, rm.sourceCodeVar, rm.sourceCodeValue)
	} else {
		rm.IncludeCode = fmt.Sprintf(excludeCode, rm.sourceCodeVar, rm.sourceCodeValue)
	}

	return nil
}
