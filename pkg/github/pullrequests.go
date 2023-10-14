// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package github

import (
	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

type createPullRequestQuery struct {
	CreatePullRequest struct {
		PullRequest struct {
			Number int
		}
	} `graphql:"createPullRequest(input: $input)"`
}

func (c *Client) CreatePullRequest(repoID githubv4.ID, baseRef, headRef, title, body string) (int, error) {
	var q createPullRequestQuery

	c.log.WithFields(logrus.Fields{
		"base": baseRef,
		"head": headRef,
	}).Debugf("CreatePullRequest()")

	input := githubv4.CreatePullRequestInput{
		RepositoryID:        repoID,
		BaseRefName:         githubv4.String(baseRef),
		HeadRefName:         githubv4.String(headRef),
		Title:               githubv4.String(title),
		Body:                githubv4.NewString(githubv4.String(body)),
		MaintainerCanModify: githubv4.NewBoolean(true),
	}

	err := c.client.Mutate(c.ctx, &q, input, nil)
	if err != nil {
		return 0, err
	}

	return q.CreatePullRequest.PullRequest.Number, nil
}

type pullRequestsQuery struct {
	Organization struct {
		Repository struct {
			PullRequests struct {
				Nodes []struct {
					Number int
				}
			} `graphql:"pullRequests(first: 1, baseRefName: $base, headRefName: $head, states: OPEN)"`
		} `graphql:"repository(name: $repo)"`
	} `graphql:"organization(login: $login)"`
}

func (c *Client) GetPullRequestForBranch(org, repo, baseRef, headRef string) (int, error) {
	variables := map[string]interface{}{
		"login": githubv4.String(org),
		"repo":  githubv4.String(repo),
		"base":  githubv4.String(baseRef),
		"head":  githubv4.String(headRef),
	}

	var q pullRequestsQuery

	c.log.WithFields(logrus.Fields{
		"org":  org,
		"repo": repo,
		"base": baseRef,
		"head": headRef,
	}).Debugf("PullRequestForBranchExists()")

	err := c.client.Query(c.ctx, &q, variables)
	if err != nil {
		return 0, err
	}

	if len(q.Organization.Repository.PullRequests.Nodes) == 0 {
		return 0, nil
	}

	return q.Organization.Repository.PullRequests.Nodes[0].Number, nil
}
