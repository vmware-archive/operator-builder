// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StandaloneSetNames(t *testing.T) {
	sharedNameInput := WorkloadShared{
		Name: "shared-name",
		Kind: "StandaloneWorkload",
	}

	sharedNameExpected := WorkloadShared{
		Name:        "shared-name",
		PackageName: "sharedname",
		Kind:        "StandaloneWorkload",
	}

	for _, tt := range []struct {
		name     string
		input    *StandaloneWorkload
		expected *StandaloneWorkload
	}{
		{
			name: "standalone workload missing root command",
			input: &StandaloneWorkload{
				WorkloadShared: sharedNameInput,
			},
			expected: &StandaloneWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: StandaloneWorkloadSpec{
					CompanionCliRootcmd: CliCommand{},
				},
			},
		},
		{
			name: "standalone workload with root command missing description",
			input: &StandaloneWorkload{
				WorkloadShared: sharedNameInput,
				Spec: StandaloneWorkloadSpec{
					API: APISpec{
						Kind: "StandaloneWorkloadTest",
					},
					CompanionCliRootcmd: CliCommand{
						Name: "hasrootcommand",
					},
				},
			},
			expected: &StandaloneWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: StandaloneWorkloadSpec{
					API: APISpec{
						Kind: "StandaloneWorkloadTest",
					},
					CompanionCliRootcmd: CliCommand{
						Name:        "hasrootcommand",
						Description: "Manage standaloneworkloadtest workload",
						VarName:     "Hasrootcommand",
						FileName:    "hasrootcommand",
					},
				},
			},
		},
		{
			name: "standalone workload with root command",
			input: &StandaloneWorkload{
				WorkloadShared: sharedNameInput,
				Spec: StandaloneWorkloadSpec{
					API: APISpec{
						Kind: "StandaloneWorkloadTest",
					},
					CompanionCliRootcmd: CliCommand{
						Name:        "hasrootcommand",
						Description: "Manage standaloneworkloadtest workload custom",
					},
				},
			},
			expected: &StandaloneWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: StandaloneWorkloadSpec{
					API: APISpec{
						Kind: "StandaloneWorkloadTest",
					},
					CompanionCliRootcmd: CliCommand{
						Name:        "hasrootcommand",
						Description: "Manage standaloneworkloadtest workload custom",
						VarName:     "Hasrootcommand",
						FileName:    "hasrootcommand",
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.SetNames()
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}
