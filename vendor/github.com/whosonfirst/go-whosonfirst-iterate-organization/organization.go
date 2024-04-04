package organization

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/whosonfirst/go-whosonfirst-github/organizations"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func init() {
	ctx := context.Background()
	emitter.RegisterEmitter(ctx, "org", NewOrganizationEmitter)
}

// type OrganizationEmitter implements the `emitter.Emitter` interface for iterating multiple repositories in a GitHub organization.
type OrganizationEmitter struct {
	emitter.Emitter
	target string
	query  url.Values
}

// NewOrganizationEmitter returns a new `OrganizationEmitter` configured by 'uri' which takes the form
// of:
//
//	org://{PATH}?{PARAMETERS}
//
// Where {PATH} is an optional path where individual Git repositories should be downloaded for processing; {PARAMETERS} is
// optional and may be any of the valid parameters used in URIs to create a new `whosonfirst/go-whosonfirst-iterate-git.GitEmitter`.
// If {PATH} is not defined then Git repositories are download in to, and processed from, memory. If {PATH} is defined any Git repositories
// downloaded will be remove after processing (unless the `?preserve=1` query parameter is present).
func NewOrganizationEmitter(ctx context.Context, uri string) (emitter.Emitter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	em := &OrganizationEmitter{
		target: u.Path,
		query:  q,
	}

	return em, nil
}

// WalkURI fetchesone or more respositories belonging to a GitHub orgnization invoking 'cb' for each file in those respositores.
// The list of files to process is determined by 'uri' which takes the form of:
//
//	{GITHUB_ORGANIZATION}://?prefix={STRING}&exclude={STRING}&access_token={STRING}
//
// Where {PREFIX} is zero or more "prefix" parameters to filter the list of repositories by for inclusion; {EXCLUDE} is zero or
// more "exclude" query parameters to filter the list of repositories by for exclusion; {ACCESS_TOKEN} is an optional GitHub API
// access token to include with the underlying calls to the GitHub API.
func (em *OrganizationEmitter) WalkURI(ctx context.Context, cb emitter.EmitterCallbackFunc, uri string) error {

	u, err := url.Parse(uri)

	if err != nil {
		return fmt.Errorf("Failed to parse URI, %w", err)
	}

	organization := u.Scheme

	q := u.Query()

	prefix := q["prefix"]
	exclude := q["exclude"]

	access_token := q.Get("access_token")

	retry := false
	max_retries := 3
	retry_after := 10 // seconds

	if q.Has("retry") {

		v, err := strconv.ParseBool(q.Get("retry"))

		if err != nil {
			return fmt.Errorf("Invalid ?retry= parameter, %w", err)
		}

		q.Del("retry")
		retry = v
	}

	if q.Has("max_retries") {

		v, err := strconv.Atoi(q.Get("max_retries"))

		if err != nil {
			return fmt.Errorf("Invalid ?max_retries= parameter, %w", err)
		}

		q.Del("max_retries")
		max_retries = v
	}

	if q.Has("retry_after") {

		v, err := strconv.Atoi(q.Get("retry_after"))

		if err != nil {
			return fmt.Errorf("Invalid ?retry_after= parameter, %w", err)
		}

		q.Del("retry_after")
		retry_after = v
	}

	list_opts := organizations.NewDefaultListOptions()
	list_opts.Prefix = prefix
	list_opts.Exclude = exclude
	list_opts.AccessToken = access_token

	repos, err := organizations.ListRepos(organization, list_opts)

	if err != nil {
		return fmt.Errorf("Failed to list repos, %w", err)
	}

	iterator_sources := make([]string, len(repos))

	for idx, repo := range repos {
		iterator_sources[idx] = fmt.Sprintf("https://github.com/%s/%s.git", organization, repo)
	}

	iter_q := url.Values{}

	for k, v_list := range em.query {

		for _, v := range v_list {
			iter_q.Set(k, v)
		}
	}

	if retry {
		iter_q.Set("_retry", strconv.FormatBool(retry))
		iter_q.Set("_max_retries", strconv.Itoa(max_retries))
		iter_q.Set("_retry_after", strconv.Itoa(retry_after))
	}

	iterator_uri := url.URL{}
	iterator_uri.Scheme = "git"
	iterator_uri.Path = em.target
	iterator_uri.RawQuery = iter_q.Encode()

	iter, err := iterator.NewIterator(ctx, iterator_uri.String(), cb)

	if err != nil {
		return fmt.Errorf("Failed to create new iterator, %w", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		return fmt.Errorf("Failed to iterate URIs, %w", err)
	}

	return nil
}
