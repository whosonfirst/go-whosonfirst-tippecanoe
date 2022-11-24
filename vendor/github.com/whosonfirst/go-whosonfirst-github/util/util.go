package util

import (
	"context"
	"github.com/google/go-github/v27/github"
	"golang.org/x/oauth2"
)

func NewClientAndContext(token string) (*github.Client, context.Context, error) {

	// https://godoc.org/github.com/google/go-github/github#Client

	client := github.NewClient(nil)
	ctx := context.TODO()

	if token != "" {

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)

		tc := oauth2.NewClient(ctx, ts)

		client = github.NewClient(tc)
	}

	return client, ctx, nil
}
