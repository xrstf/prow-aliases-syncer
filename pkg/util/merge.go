package util

import (
	"strings"

	"go.xrstf.de/prow-aliases-syncer/pkg/github"
	"go.xrstf.de/prow-aliases-syncer/pkg/prow"
)

func BuildNewOwners(old *prow.OwnersAliases, teams []github.Team) *prow.OwnersAliases {
	result := &prow.OwnersAliases{}

	for teamName := range old.Aliases {
		for _, team := range teams {
			if team.Slug == teamName {
				if result.Aliases == nil {
					result.Aliases = map[string][]string{}
				}

				for i, m := range team.Members {
					team.Members[i] = strings.ToLower(m)
				}

				result.Aliases[team.Slug] = team.Members
				break
			}
		}
	}

	return result
}
