package v1

import (
	"fmt"
	"strings"
)

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

func (fm FieldMarker) GetType() string {
	return "FieldMarker"
}

func (fm FieldMarker) ExtractSpecField() *APISpecField {
	specField := &APISpecField{
		FieldName:          strings.ToTitle(fm.Name),
		ManifestFieldName:  fm.Name,
		DataType:           fm.Type,
		APISpecContent:     getAPISpecContent(fm.Name, fm.Type),
		DocumentationLines: getFieldDescription(fm.Description),
		ZeroVal:            fm.Type.zeroValue(),
	}

	if fm.Default != nil {
		specField.setSampleAndDefault(fm.Name, fm.Default)
	} else {
		specField.setSampleAndDefault(fm.Name, fm.originalValue)
	}

	return specField
}

func getAPISpecContent(name string, fieldType FieldType) string {
	return fmt.Sprintf("%s %s `json:%q`", strings.Title(name), fieldType, name)
}

func getFieldDescription(description *string) []string {
	if description != nil {
		return strings.Split(*description, "\n")
	}

	return nil
}
