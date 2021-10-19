// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

// OwnershipRule contains the info needed to create the controller ownership
// functionality when setting up the controller with the manager.  This allows
// the controller to reconcile the state of a deleted resource that it manages.
type OwnershipRule struct {
	Version string
	Kind    string
	CoreAPI bool
}

type OwnershipRules []OwnershipRule

func (or *OwnershipRules) addOrUpdateOwnership(version, kind, group string) {
	// determine group and kind for ownership rule generation
	newOwnershipRule := OwnershipRule{
		Version: version,
		Kind:    kind,
		CoreAPI: isCoreAPI(group),
	}

	if !or.versionKindRecorded(&newOwnershipRule) {
		*or = append(*or, newOwnershipRule)
	}
}

func (or *OwnershipRules) versionKindRecorded(newOwnershipRule *OwnershipRule) bool {
	for _, r := range *or {
		if r.Version == newOwnershipRule.Version && r.Kind == newOwnershipRule.Kind {
			return true
		}
	}

	return false
}
