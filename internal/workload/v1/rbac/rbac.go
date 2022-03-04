// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
)

// rbacWorkloadProcessor is an interface which implements processing of rbac rules
// for individual workloads (e.g. standalone, collection, component).
type rbacWorkloadProcessor interface {
	IsComponent() bool

	GetDomain() string
	GetAPIGroup() string
	GetAPIVersion() string
	GetAPIKind() string
}

// rbacRuleProcessor is an interface which implements processing of individual
// rbac rules.
type rbacRuleProcessor interface {
	addTo(*Rules)
}

const (
	coreRBACGroup = "core"
)

// ForManifest will return a set of rules for a particular manifest.  This includes
// a rule for the manifest itself, in addition to adding particular rules for whatever
// roles and cluster roles are requesting.  This is because the controller needs to have
// permissions to manage the children that roles and cluster roles are requesting.
func ForManifest(manifest *unstructured.Unstructured) (*Rules, error) {
	rules := &Rules{}

	if err := rules.addForManifest(manifest); err != nil {
		return rules, err
	}

	return rules, nil
}

// ForWorkloads will return a set of rules for a particular set of workloads.  It should be noted that
// this only returns the specific rules for the actual workload and not the managed resources.  See
// ForManifest for details on the rules for a particular manifest.
func ForWorkloads(workloads ...rbacWorkloadProcessor) *Rules {
	rules := &Rules{}

	// for each of the workloads passed in, add a rule to the set of rules
	for i := range workloads {
		rules.addForWorkload(workloads[i])
	}

	return rules
}

// GetVersionGroup is a helper function that returns a version and group given
// an apiVersion as a string.
func GetVersionGroup(apiVersion string) (version, group string) {
	apiVersionElements := strings.Split(apiVersion, "/")

	if len(apiVersionElements) == 1 {
		version = apiVersionElements[0]
		group = coreRBACGroup
	} else {
		version = apiVersionElements[1]
		group = rbacGroupFromGroup(apiVersionElements[0])
	}

	return version, group
}

// defaultResourceVerbs is a helper function to define the default verbs that are allowed
// for resources that are managed by the scaffolded controller.
func defaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
}

// defaultStatusVerbs is a helper function to define the default verbs which get placed
// onto resources that have a /status suffix.
func defaultStatusVerbs() []string {
	return []string{
		"get", "update", "patch",
	}
}

// knownIrregulars is a helper function to define known irregular kinds and their
// expected formats.
func knownIrregulars() map[string]string {
	return map[string]string{
		"resourcequota": "resourcequotas",
	}
}

func rbacGroupFromGroup(group string) string {
	if group == "" {
		return coreRBACGroup
	}

	return group
}

func rbacFieldsToString(verbs []string) string {
	return strings.Join(verbs, ";")
}

func getResourceForRBAC(kind string) string {
	rbacResource := strings.Split(kind, "/")

	if rbacResource[0] == "*" {
		kind = "*"
	} else {
		kind = getPluralRBAC(rbacResource[0])
	}

	if len(rbacResource) > 1 {
		kind = fmt.Sprintf("%s/%s", kind, rbacResource[1])
	}

	return kind
}

// getPluralRBAC will transform known irregulars into a proper type for rbac
// rules.
func getPluralRBAC(kind string) string {
	plural := resource.RegularPlural(kind)

	if knownIrregulars()[plural] != "" {
		return knownIrregulars()[plural]
	}

	return plural
}

func valueFromInterface(in interface{}, key string) (out interface{}) {
	switch asType := in.(type) {
	case map[interface{}]interface{}:
		out = asType[key]
	case map[string]interface{}:
		out = asType[key]
	case map[interface{}][]interface{}:
		out = asType[key]
	case map[string][]interface{}:
		out = asType[key]
	}

	return out
}
