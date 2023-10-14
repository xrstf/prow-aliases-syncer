// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Client struct {
	ctx    context.Context
	client *githubv4.Client
	log    logrus.FieldLogger
}

func NewClient(ctx context.Context, log logrus.FieldLogger, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	return &Client{
		ctx:    ctx,
		client: client,
		log:    log,
	}, nil
}
