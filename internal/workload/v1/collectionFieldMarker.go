package v1

import (
	"fmt"
	"strings"
)

type CollectionFieldMarker FieldMarker

func (cfm CollectionFieldMarker) String() string {
	return fmt.Sprintf("CollectionFieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		cfm.Name,
		cfm.Type,
		*cfm.Description,
		cfm.Default,
	)
}

func (cfm CollectionFieldMarker) GetType() string {
	return "CollectionFieldMarker"
}

func (cfm CollectionFieldMarker) ExtractSpecField() *APISpecField {
	specField := &APISpecField{
		FieldName:          strings.ToTitle(cfm.Name),
		ManifestFieldName:  cfm.Name,
		DataType:           cfm.Type,
		APISpecContent:     getAPISpecContent(cfm.Name, cfm.Type),
		DocumentationLines: getFieldDescription(cfm.Description),
		ZeroVal:            cfm.Type.zeroValue(),
	}

	if cfm.Default != nil {
		specField.setSampleAndDefault(cfm.Name, cfm.Default)
	} else {
		specField.setSampleAndDefault(cfm.Name, cfm.originalValue)
	}

	return specField
}
