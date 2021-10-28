// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
)

var ErrUnableToParseFieldType = errors.New("unable to parse field")

type FieldType int

const (
	FieldUnknownType FieldType = iota
	FieldString
	FieldInt
	FieldBool
	FieldSliceString
	FieldSliceInt
	FieldSliceBool
	FieldStruct
)

func (f *FieldType) UnmarshalMarkerArg(in string) error {
	types := map[string]FieldType{
		"":         FieldUnknownType,
		"string":   FieldString,
		"int":      FieldInt,
		"bool":     FieldBool,
		"[]string": FieldSliceString,
		"[]int":    FieldSliceInt,
		"[]bool":   FieldSliceBool,
	}

	if t, ok := types[in]; ok {
		if t == FieldUnknownType {
			return fmt.Errorf("%w, %s into FieldType", ErrUnableToParseFieldType, in)
		}

		*f = t

		return nil
	}

	return fmt.Errorf("%w, %s into FieldType", ErrUnableToParseFieldType, in)
}

func (f FieldType) String() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
		FieldSliceString: "[]string",
		FieldSliceInt:    "[]int",
		FieldSliceBool:   "[]bool",
		FieldStruct:      "struct",
	}

	return types[f]
}
