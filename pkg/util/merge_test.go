package util

import (
	"fmt"
	"testing"

	"go.xrstf.de/prow-aliases-syncer/pkg/github"
	"go.xrstf.de/prow-aliases-syncer/pkg/prow"

	"github.com/go-test/deep"
)

func TestBuildNewOwners(t *testing.T) {
	testcases := []struct {
		oldData  prow.OwnersAliases
		teams    []github.Team
		keep     bool
		expected prow.OwnersAliases
	}{
		{
			oldData: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"1", "2", "3"},
					"b": {"4", "5", "6"},
					"c": {"7", "8", "9"},
				},
			},
			keep:     false,
			teams:    nil,
			expected: prow.OwnersAliases{},
		},
		{
			oldData: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"3", "1", "2"},
					"b": {"4", "5", "6"},
					"c": {"7", "8", "9"},
				},
			},
			keep: false,
			teams: []github.Team{
				{
					Slug:    "a",
					Members: []string{"1", "3"},
				},
			},
			expected: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"1", "3"},
				},
			},
		},
		{
			oldData: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"1", "2", "3"},
					"b": {"4", "5", "6"},
					"c": {"7", "8", "9"},
				},
			},
			keep: true,
			teams: []github.Team{
				{
					Slug:    "a",
					Members: []string{"1", "3"},
				},
			},
			expected: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"1", "3"},
					"b": {"4", "5", "6"},
					"c": {"7", "8", "9"},
				},
			},
		},
		{
			oldData: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"1", "2", "3"},
					"b": {"4", "5", "6"},
					"c": {"7", "8", "9"},
				},
			},
			keep:  true,
			teams: nil,
			expected: prow.OwnersAliases{
				Aliases: map[string][]string{
					"a": {"1", "2", "3"},
					"b": {"4", "5", "6"},
					"c": {"7", "8", "9"},
				},
			},
		},
	}

	for i, testcase := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			result := BuildNewOwners(&testcase.oldData, testcase.teams, testcase.keep)

			if diff := deep.Equal(*result, testcase.expected); diff != nil {
				t.Fatalf("not equal: %v", diff)
			}
		})
	}
}
