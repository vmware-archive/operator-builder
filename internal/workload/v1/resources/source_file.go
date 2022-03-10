// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

// SourceFile represents a golang source code file that contains one or more
// child resource objects.
type SourceFile struct {
	Filename string
	Children []Child
}

// NewSourceFile creates a new source code file given the file path of an input manifest.
func NewSourceFile(manifest *Manifest) *SourceFile {
	return &SourceFile{
		Filename: uniqueName(manifest),
	}
}

// uniqueName returns the unique file name for a source file.
func uniqueName(manifest *Manifest) (name string) {
	name = filepath.Clean(manifest.RelativeFileName)
	name = strings.ReplaceAll(name, "/", "_")               // get filename from path
	name = strings.ReplaceAll(name, filepath.Ext(name), "") // strip ".yaml"
	name = strings.ReplaceAll(name, ".", "")                // strip "." e.g. hidden files
	name += ".go"                                           // add correct file ext
	name = utils.ToFileName(name)                           // kebab-case to snake_case

	// strip any prefix that begins with _ or even multiple _s because go does not recognize these files
	for _, char := range name {
		if string(char) == "_" {
			name = strings.TrimPrefix(name, "_")
		} else {
			break
		}
	}

	return name
}

// GetFuncNames returns the create and init function names to be used in the scaffolded
// source files.
func GetFuncNames(sourceFiles []SourceFile) (createFuncNames, initFuncNames []string) {
	for _, sourceFile := range sourceFiles {
		for i := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", sourceFile.Children[i].UniqueName)

			if strings.EqualFold(sourceFile.Children[i].Kind, "customresourcedefinition") {
				initFuncNames = append(initFuncNames, funcName)
			}

			createFuncNames = append(createFuncNames, funcName)
		}
	}

	return createFuncNames, initFuncNames
}
