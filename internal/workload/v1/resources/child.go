// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"errors"
	"fmt"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
)

var (
	ErrChildResourceResourceMarkerInspect = errors.New("error inspecting resource markers for child resource")
	ErrChildResourceResourceMarkerProcess = errors.New("error processing resource markers for child resource")
)

// Child contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manifests.  They can be thought
// of as the individual resources which are managed by the controller during
// reconciliation and represent all resources which are passed in via the `spec.resources`
// field of the workload configuration.
type Child struct {
	Name          string
	UniqueName    string
	Group         string
	Version       string
	Kind          string
	StaticContent string
	SourceCode    string
	IncludeCode   string
}

//nolint:gocritic // needed to satisfy the stringer interface
func (child Child) String() string {
	return fmt.Sprintf(
		"{Group: %s, Version: %s, Kind: %s, Name: %s}",
		child.Group, child.Version, child.Kind, child.Name,
	)
}

func (child *Child) ProcessResourceMarkers(markerCollection *markers.MarkerCollection) error {
	// obtain the marker results from the child resource input yaml
	_, markerResults, err := markers.InspectForYAML([]byte(child.StaticContent), markers.ResourceMarkerType)
	if err != nil {
		return fmt.Errorf("%w; %s for child resource %s", err, ErrChildResourceResourceMarkerInspect, child)
	}

	// ensure we have the expected number of resource markers
	//   - 0: return immediately as resource markers are not required
	//   - 1: continue processing normally
	//   - 2: return an error notifying the user that we only expect 1
	//        resource marker
	if len(markerResults) == 0 {
		return nil
	}

	//nolint: godox // depends on https://github.com/vmware-tanzu-labs/operator-builder/issues/271
	// TODO: we need to ensure only one marker is found and return an error if we find more than one.
	// this becomes difficult as the results are returned as yaml nodes.  for now, we just focus on the
	// first result and all others are ignored but we should notify the user.
	result := markerResults[0]

	// process the marker
	marker, ok := result.Object.(markers.ResourceMarker)
	if !ok {
		return ErrChildResourceResourceMarkerProcess
	}

	if err := marker.Process(markerCollection); err != nil {
		return fmt.Errorf("%w; %s for child resource %s", err, ErrChildResourceResourceMarkerProcess, child)
	}

	if marker.GetIncludeCode() != "" {
		child.IncludeCode = marker.GetIncludeCode()
	}

	return nil
}
