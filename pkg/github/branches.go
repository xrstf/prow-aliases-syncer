package github

import (
	"sort"
	"strings"
	"time"

	"go.xrstf.de/prow-aliases-syncer/pkg/prow"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
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
								File          struct {
									Object struct {
										Blob struct {
											Text string
										} `graphql:"... on Blob"`
									}
								} `graphql:"file(path: $filename)"`
							} `graphql:"... on Commit"`
						}
					}
				} `graphql:"refs(first: 100, orderBy: {field: ALPHABETICAL, direction: ASC}, refPrefix: $prefix)"`
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"repositories(first: 20, orderBy: {field: NAME, direction: ASC}, after: $cursor)"`
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

func (c *Client) GetRepositoriesAndBranches(org string) ([]Repository, error) {
	result := []Repository{}
	cursor := ""

	for {
		var (
			items []Repository
			err   error
		)

		items, cursor, err = c.getRepositoriesAndBranches(org, cursor)
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

func (c *Client) getRepositoriesAndBranches(org string, cursor string) ([]Repository, string, error) {
	variables := map[string]interface{}{
		"filename": githubv4.String(prow.OwnersAliasesFilename),
		"login":    githubv4.String(org),
		"prefix":   githubv4.String("refs/heads/"),
		"cursor":   (*githubv4.String)(nil),
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

	for _, r := range q.Organization.Repositories.Nodes {
		repo := Repository{
			ID:       r.ID,
			Name:     r.Name,
			Branches: []Branch{},
		}

		for _, b := range r.Refs.Nodes {
			repo.Branches = append(repo.Branches, Branch{
				Name:             b.Name,
				MostRecentCommit: b.Target.Commit.CommittedDate.Time,
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
