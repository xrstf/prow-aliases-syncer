// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package github

import (
	"sort"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"go.xrstf.de/prow-aliases-syncer/pkg/prow"

	"k8s.io/apimachinery/pkg/util/sets"
)

type repositoriesBranchesQuery struct {
	Organization struct {
		Repositories struct {
			Nodes []struct {
				ID   githubv4.ID
				Name string
				Refs struct {
					Nodes []struct {
						Name   string
						Target struct {
							Commit struct {
								OID           string
								CommittedDate githubv4.DateTime

								// fetch the current state of the OWNERS_ALIASES file
								File struct {
									Object struct {
										Blob struct {
											Text string
										} `graphql:"... on Blob"`
									}
								} `graphql:"file(path: $filename)"`

								// fetch the most recent history for this branch, so we
								// can check if the activity was only caused by us updating
								// the owners file, or if there are other commits in here
								History struct {
									Nodes []struct {
										CommittedDate githubv4.DateTime
										Author        struct {
											User struct {
												Login string
											}
										}
									}
								} `graphql:"history(first: $peek)"`
							} `graphql:"... on Commit"`
						}
					}
				} `graphql:"refs(first: 100, orderBy: {field: ALPHABETICAL, direction: ASC}, refPrefix: $prefix)"`
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"repositories(first: 5, orderBy: {field: NAME, direction: ASC}, after: $cursor)"`
	} `graphql:"organization(login: $login)"`
}

type Repository struct {
	ID       githubv4.ID
	Name     string
	Branches []Branch
}

type Branch struct {
	Name             string
	MostRecentCommit time.Time
	Aliases          string
}

func (c *Client) GetRepositoriesAndBranches(org string, ignoredUsers []string, peekDepth int) ([]Repository, error) {
	result := []Repository{}
	cursor := ""

	// secret optimization: if no users are ignored (this should never happen,
	// as you should always ignore the bot who runs this tool), there is no need
	// to peek into any commits, we can just take the commitDate from the latest
	// commit
	if len(ignoredUsers) == 0 {
		peekDepth = 0
	}

	for {
		var (
			items []Repository
			err   error
		)

		items, cursor, err = c.getRepositoriesAndBranches(org, ignoredUsers, peekDepth, cursor)
		if err != nil {
			return nil, err
		}

		result = append(result, items...)

		if cursor == "" {
			break
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

func (c *Client) getRepositoriesAndBranches(org string, ignoredUsers []string, peekDepth int, cursor string) ([]Repository, string, error) {
	variables := map[string]interface{}{
		"filename": githubv4.String(prow.OwnersAliasesFilename),
		"login":    githubv4.String(org),
		"prefix":   githubv4.String("refs/heads/"),
		"cursor":   (*githubv4.String)(nil),
		"peek":     githubv4.Int(peekDepth),
	}

	if cursor != "" {
		variables["cursor"] = githubv4.String(cursor)
	}

	var q repositoriesBranchesQuery

	c.log.WithFields(logrus.Fields{
		"org":    org,
		"cursor": string(cursor),
	}).Debugf("GetRepositoriesAndBranches()")

	// igore errors because a missing file in a branch would cause an error
	// and we have no option to introspect the error
	_ = c.client.Query(c.ctx, &q, variables)
	// if err != nil {
	// 	return nil, "", err
	// }

	result := []Repository{}
	ignored := sets.NewString(ignoredUsers...)

	for _, r := range q.Organization.Repositories.Nodes {
		repo := Repository{
			ID:       r.ID,
			Name:     r.Name,
			Branches: []Branch{},
		}

		for _, b := range r.Refs.Nodes {
			// if the following loop finds no commit (e.g. because we ignore all
			// relevant users), we want to assume that the branch is "alive" and
			// needs updating, so that we fail safely (i.e. branches do not get lost
			// because we didn't peek far enough into their history)
			mostRecentCommit := b.Target.Commit.CommittedDate.Time

			// look through the most recent N commits and find the most recent one,
			// while ignoring a certain group of users (i.e. do not count the commits
			// that this tool is producing)
			for _, c := range b.Target.Commit.History.Nodes {
				if !ignored.Has(c.Author.User.Login) {
					mostRecentCommit = c.CommittedDate.Time
					break
				}
			}

			repo.Branches = append(repo.Branches, Branch{
				Name:             b.Name,
				MostRecentCommit: mostRecentCommit,
				Aliases:          b.Target.Commit.File.Object.Blob.Text,
			})
		}

		sort.Slice(repo.Branches, func(i, j int) bool {
			return strings.ToLower(repo.Branches[i].Name) < strings.ToLower(repo.Branches[j].Name)
		})

		result = append(result, repo)
	}

	newCursor := ""
	if q.Organization.Repositories.PageInfo.HasNextPage {
		newCursor = string(q.Organization.Repositories.PageInfo.EndCursor)
	}

	return result, newCursor, nil
}
