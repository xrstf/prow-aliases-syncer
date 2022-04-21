package github

import (
	"sort"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// organization.teams do not support pagination, neither does the members collection
type teamMembersQuery struct {
	Organization struct {
		Teams struct {
			Nodes []struct {
				Name    string
				Members struct {
					Nodes []struct {
						Login string
					}
				} `graphql:"members(first: 100, orderBy: {field: LOGIN, direction: ASC})"`
			}
		} `graphql:"teams(first: 100, orderBy: {field: NAME, direction: ASC})"`
	} `graphql:"organization(login: $login)"`
}

type Team struct {
	Slug    string
	Members []string
}

func (c *Client) GetTeams(org string) ([]Team, error) {
	variables := map[string]interface{}{
		"login": githubv4.String(org),
	}

	var q teamMembersQuery

	c.log.WithFields(logrus.Fields{
		"org": org,
	}).Debugf("GetTeams()")

	err := c.client.Query(c.ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	result := []Team{}
	for _, t := range q.Organization.Teams.Nodes {
		members := []string{}
		for _, m := range t.Members.Nodes {
			members = append(members, m.Login)
		}

		result = append(result, Team{
			Slug:    t.Name,
			Members: members,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Slug) < strings.ToLower(result[j].Slug)
	})

	return result, nil
}
