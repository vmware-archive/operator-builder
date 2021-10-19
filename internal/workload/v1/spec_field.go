// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
	"strings"
)

var ErrOverwriteExistingValue = errors.New("an attempt to overwrite existing value was made")

// APISpecField represents a single field in a custom API type.
type APISpecField struct {
	FieldName          string
	Path               string
	ManifestFieldName  string
	DataType           FieldType
	DefaultVal         string
	ZeroVal            string
	APISpecContent     string
	SampleField        string
	DocumentationLines []string
}

type APISpecFields []APISpecField

func (api *APISpecField) setSampleAndDefault(name string, value interface{}) {
	if api.DataType == FieldString {
		api.DefaultVal = fmt.Sprintf("%q", value)
		api.SampleField = fmt.Sprintf("%s: %q", name, value)
	} else {
		api.DefaultVal = fmt.Sprintf("%v", value)
		api.SampleField = fmt.Sprintf("%s: %v", name, value)
	}
}

func (api *APISpecField) BuildMap(output map[string]interface{}) error {
	obj := output
	parts := strings.Split(api.Path, ".")
	last := parts[len(parts)-1]

	for _, part := range parts[:len(parts)-1] {
		if _, ok := obj[part].(map[string]interface{}); !ok {
			if obj[part] != nil {
				return fmt.Errorf("%w at %s", ErrOverwriteExistingValue, api.Path)
			}
			obj[part] = make(map[string]interface{})
		}

		//nolint: forcetypeassert // type checking occurs above
		obj = obj[part].(map[string]interface{})
	}

	if obj[last] != nil {
		return fmt.Errorf("%w at %s", ErrOverwriteExistingValue, api.Path)
	}

	obj[last] = api.DefaultVal

	return nil
}

func (apis *APISpecFields) BuildMaps(output map[string]interface{}) error {
	a := *apis

	for i := range a {
		err := a[i].BuildMap(output)
		if err != nil {
			return err
		}
	}

	return nil
}
