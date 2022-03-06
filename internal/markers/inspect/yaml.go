// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package inspect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/parser"
)

type YAMLResult struct {
	*parser.Result
	Nodes []*yaml.Node
}

func (s *Inspector) InspectYAML(data []byte, transforms ...YAMLTransformer) ([]*yaml.Node, []*YAMLResult, error) {
	var nodes []*yaml.Node

	yamlDecoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var node yaml.Node

		if err := yamlDecoder.Decode(&node); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, nil, fmt.Errorf("error unmarshaling yaml, %w", err)
		}

		nodes = append(nodes, &node)
	}

	var results []*YAMLResult

	for _, node := range nodes {
		docResults := s.inspectYAML(node)

		results = append(results, docResults...)
	}

	for _, result := range results {
		if v, ok := result.Result.Object.(error); ok {
			return nodes, results, v
		}
	}

	for _, transform := range transforms {
		if err := transform(results...); err != nil {
			return nodes, nil, err
		}
	}

	return nodes, results, nil
}

func (s *Inspector) inspectYAML(nodes ...*yaml.Node) (results []*YAMLResult) {
	for _, node := range nodes {
		results = append(results, s.inspectYAMLComments(node)...)

		if node.Kind == yaml.MappingNode {
			results = append(results, s.inspectYAMLMap(node.Content...)...)
		} else if node.Content != nil {
			results = append(results, s.inspectYAML(node.Content...)...)
		}
	}

	return results
}

func (s *Inspector) inspectYAMLMap(nodes ...*yaml.Node) (results []*YAMLResult) {
	for i := 0; i < len(nodes); i += 2 {
		results = append(results, s.inspectYAMLComments(nodes[i], nodes[i+1])...)

		if nodes[i+1].Kind == yaml.MappingNode {
			results = append(results, s.inspectYAMLMap(nodes[i+1].Content...)...)
		} else {
			results = append(results, s.inspectYAML(nodes[i+1].Content...)...)
		}
	}

	return results
}

func (s *Inspector) inspectYAMLComments(nodes ...*yaml.Node) (results []*YAMLResult) {
	var markers []*parser.Result

	for _, node := range nodes {
		// hack: the parse method should theoretically handle proper parsing, however we found situations in which
		// other markers (e.g. kubebuilder markers) conflict with our own defined markers which return an error.  This
		// is a quick fix for now until we figure out a way to avoid parsing comflicts.
		if !s.hasMarkerText(node) {
			continue
		}

		markers = append(markers, s.parse(fmt.Sprintf("%s\n%s\n%s", node.HeadComment, node.LineComment, node.FootComment))...)
	}

	for _, marker := range markers {
		result := &YAMLResult{
			Result: marker,
			Nodes:  nodes,
		}

		results = append(results, result)
	}

	return results
}

func (s *Inspector) hasMarkerText(node *yaml.Node) bool {
	for _, markerText := range s.Registry.MarkerStrings() {
		switch {
		case strings.Contains(node.HeadComment, markerText):
			return true
		case strings.Contains(node.FootComment, markerText):
			return true
		case strings.Contains(node.LineComment, markerText):
			return true
		}
	}

	return false
}
