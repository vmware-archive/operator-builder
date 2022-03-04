// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"fmt"
)

// Rule contains the info needed to create the kubebuilder:rbac markers in
// the controller.
type Rule struct {
	Group    string
	Resource string
	URLs     []string
	Verbs    []string
}

// ToMarker will return a specific marker in string format.
func (rule *Rule) ToMarker() string {
	if len(rule.URLs) > 0 {
		return fmt.Sprintf("%s:verbs=%s,urls=%s",
			kubebuilderPrefix,
			getFieldString(rule.Verbs),
			getFieldString(rule.URLs),
		)
	}

	return fmt.Sprintf("%s:groups=%s,resources=%s,verbs=%s",
		kubebuilderPrefix,
		rule.Group,
		rule.Resource,
		getFieldString(rule.Verbs),
	)
}

// addTo satisfies the rbacRuleProcessor interface by defining the logic that adds a rule into an
// existing set of rules.
func (rule *Rule) addTo(rules *Rules) {
	rs := *rules

	if !rs.hasRule(rule) {
		*rules = append(*rules, *rule)
	} else {
		for i := range rs {
			if rs[i].groupResourceEqual(rule) {
				for _, verb := range rule.Verbs {
					rs[i].addVerb(verb)
				}
			}
		}
	}
}

// addVerb adds a verb to an existing rule.  The verb is only added if it is not
// found to prevent duplication of markers that are generated in the controller.
func (rule *Rule) addVerb(verb string) {
	var found bool

	for _, existingVerb := range rule.Verbs {
		if existingVerb == verb {
			found = true

			break
		}
	}

	if !found {
		rule.Verbs = append(rule.Verbs, verb)
	}
}

// groupResourceEqual determines if the group and resource are equal given an
// input rule.
func (rule *Rule) groupResourceEqual(compared *Rule) bool {
	if rule.Group == compared.Group && rule.Resource == compared.Resource {
		return true
	}

	return false
}
