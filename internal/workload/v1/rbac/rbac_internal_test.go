// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func NewRBACTest() *RBACRule {
// 	return &RBACRule{
// 		Group:    "core",
// 		Resource: "exampleResource",
// 		Verbs:    []string{"get"},
// 	}
// }

// func TestRBACRule_AddVerb(t *testing.T) {
// 	t.Parallel()

// 	type args struct {
// 		verb string
// 	}

// 	tests := []struct {
// 		name string
// 		rbac *RBACRule
// 		args args
// 	}{
// 		{
// 			name: "Test adding new verb",
// 			rbac: NewRBACTest(),
// 			args: args{
// 				verb: "delete",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			r := tt.rbac
// 			r.addVerb(tt.args.verb)

// 			assert.Equal(t, r.Verbs, []string{"get", "delete"})
// 		})
// 	}
// }
