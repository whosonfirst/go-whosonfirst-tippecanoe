package github

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"

	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v3"

	"github.com/whosonfirst/go-whosonfirst-github/organizations"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	wof_uri "github.com/whosonfirst/go-whosonfirst-uri"
)

func init() {
	ctx := context.Background()
	iterate.RegisterIterator(ctx, "githuborg", NewOrganizationIterator)
}

// type OrganizationIterator implements the `iterator.Iterator` interface for iterating multiple repositories in a GitHub organization.
type OrganizationIterator struct {
	iterate.Iterator
	target string
	query  url.Values
	dedupe bool
	lookup *sync.Map
	// The count of documents that have been processed so far.
	seen int64
	// Boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewOrganizationIterator returns a new `OrganizationIterator` configured by 'uri' which takes the form
// of:
//
//	org://{PATH}?{PARAMETERS}
//
// Where {PATH} is an optional path where individual Git repositories should be downloaded for processing; {PARAMETERS} is
// optional and may be any of the valid parameters used in URIs to create a new `whosonfirst/go-whosonfirst-iterate-git.GitIterator`.
// If {PATH} is not defined then Git repositories are download in to, and processed from, memory. If {PATH} is defined any Git repositories
// downloaded will be remove after processing (unless the `?preserve=1` query parameter is present).
func NewOrganizationIterator(ctx context.Context, uri string) (iterate.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	it := &OrganizationIterator{
		target:    u.Path,
		query:     q,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	// Note this is "?dedupe=" and not "?_dedupe=" which is handled in
	// go-whosonfirst-iterate/iterator.NewIterator. This package has its
	// own "?dedupe=" flag because we create a new iterator instance for
	// each iterator source (which is a list of repos in an organization)
	// and we want to deduplicate records across iterators.

	if q.Has("dedupe") {

		v, err := strconv.ParseBool(q.Get("dedupe"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '?dedupe=' parameter, %w", err)
		}

		if v {
			it.lookup = new(sync.Map)
			it.dedupe = v
		}

	}

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
// The list of files to process is determined by 'uris' which takes the form of:
//
//	{GITHUB_ORGANIZATION}://?prefix={STRING}&exclude={STRING}&access_token={STRING}
//
// Where {PREFIX} is zero or more "prefix" parameters to filter the list of repositories by for inclusion; {EXCLUDE} is zero or
// more "exclude" query parameters to filter the list of repositories by for exclusion; {ACCESS_TOKEN} is an optional GitHub API
// access token to include with the underlying calls to the GitHub API.
func (it *OrganizationIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*iterate.Record, error] {

	return func(yield func(rec *iterate.Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			logger := slog.Default()
			logger = logger.With("uri", uri)

			u, err := url.Parse(uri)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to parse URI, %w", err)) {
					return
				}

				continue
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
					if !yield(nil, fmt.Errorf("Invalid ?retry= parameter, %w", err)) {
						return
					}

					continue
				}

				q.Del("retry")
				retry = v
			}

			if q.Has("max_retries") {

				v, err := strconv.Atoi(q.Get("max_retries"))

				if err != nil {
					if !yield(nil, fmt.Errorf("Invalid ?max_retries= parameter, %w", err)) {
						return
					}

					continue
				}

				q.Del("max_retries")
				max_retries = v
			}

			if q.Has("retry_after") {

				v, err := strconv.Atoi(q.Get("retry_after"))

				if err != nil {
					if !yield(nil, fmt.Errorf("Invalid ?retry_after= parameter, %w", err)) {
						return
					}

					continue
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
				if !yield(nil, fmt.Errorf("Failed to list repos, %w", err)) {
					return
				}

				continue
			}

			iterator_sources := make([]string, len(repos))

			for idx, repo := range repos {
				iterator_sources[idx] = fmt.Sprintf("https://github.com/%s/%s.git", organization, repo)
			}

			//

			iter_q := url.Values{}

			for k, v_list := range it.query {

				for _, v := range v_list {
					iter_q.Set(k, v)
				}
			}

			if retry {
				iter_q.Set("_retry", strconv.FormatBool(retry))
				iter_q.Set("_max_retries", strconv.Itoa(max_retries))
				iter_q.Set("_retry_after", strconv.Itoa(retry_after))
			}

			// To do: Add support for go-whosonfirst-iterate-github
			// https://github.com/whosonfirst/go-whosonfirst-iterate-organization/issues/2

			iterator_uri := url.URL{}
			iterator_uri.Scheme = "git"
			iterator_uri.Path = it.target
			iterator_uri.RawQuery = iter_q.Encode()

			iter, err := iterate.NewIterator(ctx, iterator_uri.String())

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to create new iterator, %w", err)) {
					return
				}

				continue
			}

			for rec, err := range iter.Iterate(ctx, iterator_sources...) {

				if err != nil {
					if !yield(nil, err) {
						return
					}
					continue
				}

				if it.dedupe {

					id, uri_args, err := wof_uri.ParseURI(rec.Path)

					if err != nil {
						if !yield(nil, fmt.Errorf("Failed to parse %s, %w", rec.Path, err)) {
							return
						}

						continue
					}

					rel_path, err := wof_uri.Id2RelPath(id, uri_args)

					if err != nil {
						if !yield(nil, fmt.Errorf("Failed to derive relative path for %s, %w", rec.Path, err)) {
							return
						}

						continue
					}

					_, exists := it.lookup.LoadOrStore(rel_path, true)

					if exists {
						slog.Debug("Skip record because duplicate", "path", rel_path)
						continue
					}
				}

				if !yield(rec, nil) {
					return
				}
			}
		}
	}
}

// Seen() returns the total number of records processed so far.
func (it *OrganizationIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *OrganizationIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *OrganizationIterator) Close() error {
	return nil
}
