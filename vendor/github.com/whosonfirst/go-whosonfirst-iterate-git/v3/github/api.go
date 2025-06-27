package github

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/go-github/v72/github"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
	"golang.org/x/oauth2"
)

func init() {
	ctx := context.Background()
	err := iterate.RegisterIterator(ctx, "githubapi", NewGitHubAPIIterator)

	if err != nil {
		panic(err)
	}
}

// GitHubIterator implements the `Iterator` interface for crawling records in a GitHub repository using the GitHub API.
type GitHubAPIIterator struct {
	iterate.Iterator
	// The owner or organization of the GitHub repo being iterated over.
	owner string
	// The name of the GitHub repositoty being iterated over.
	repo string
	// The branch of the GitHub repositoty being iterated over. (Default is main.)
	branch string
	// Boolean flag indicating whether iteration should be done concurrently. (Default is false.)
	concurrent bool
	// The GitHub API client.
	client *github.Client
	// Time throttle for limiting API requests.
	throttle <-chan time.Time
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// The count of documents that have been processed so far.
	seen int64
	// Boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewGitHubAPIIterator() returns a new `GitHubAPIIterator` instance configured by 'uri' in the form of:
//
//	githubapi://{ORGANIZATION}/{REPO}/{BRANCH}?{PARAMETERS}
//
// Where {ORGANIZATION} is the name of the owner (organization) of the repository being iterated over and {REPO} is the name of that repository.
// {BRANCH} is an optional path to indicate a specific branch to use (default is "main"). {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
// * `?concurrent=` A boolean value indicating whether the repository contents should be iterated over concurrently.
func NewGitHubAPIIterator(ctx context.Context, uri string) (iterate.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	rate := time.Second / 10
	throttle := time.Tick(rate)

	it := &GitHubAPIIterator{
		throttle:  throttle,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	it.owner = u.Host

	path := strings.TrimLeft(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 1 {
		return nil, errors.New("Invalid path")
	}

	it.repo = parts[0]
	it.branch = DEFAULT_BRANCH

	q := u.Query()

	token := q.Get("access_token")
	branch := q.Get("branch")

	if token == "" {
		return nil, errors.New("Missing access token")
	}

	if branch != "" {
		it.branch = branch
	}

	if q.Has("concurrent") {

		c, err := strconv.ParseBool(q.Get("concurrent"))

		if err != nil {
			return nil, err
		}

		it.concurrent = c
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	it.client = client

	f, err := filters.NewQueryFiltersFromQuery(ctx, q)

	if err != nil {
		return nil, err
	}

	it.filters = f

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *GitHubAPIIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*iterate.Record, error] {

	return func(yield func(rec *iterate.Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			logger := slog.Default()
			logger = logger.With("repo", it.repo)
			logger = logger.With("uri", uri)

			file_contents, dir_contents, _, err := it.client.Repositories.GetContents(ctx, it.owner, it.repo, uri, nil)

			if err != nil {

				logger.Warn("Failed to get repository contents", "error", err)

				if !yield(nil, err) {
					return
				}

				continue
			}

			if file_contents != nil {

				atomic.AddInt64(&it.seen, 1)

				rec, err := it.walkFileContents(ctx, file_contents)

				if err != nil {
					logger.Warn("Failed to get file contents", "error", err)
				}

				if !yield(rec, err) {
					return
				}

				continue
			}

			if dir_contents != nil {

				if it.concurrent {
					for rec, err := range it.walkDirectoryContentsConcurrently(ctx, dir_contents) {
						if !yield(rec, err) {
							return
						}
					}
				} else {

					for rec, err := range it.walkDirectoryContents(ctx, dir_contents) {
						if !yield(rec, err) {
							return
						}
					}

				}
			}
		}
	}

}

// Seen() returns the total number of records processed so far.
func (it *GitHubAPIIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *GitHubAPIIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *GitHubAPIIterator) Close() error {
	return nil
}

func (it *GitHubAPIIterator) walkDirectoryContents(ctx context.Context, contents []*github.RepositoryContent) iter.Seq2[*iterate.Record, error] {

	return func(yield func(rec *iterate.Record, err error) bool) {

		for _, e := range contents {

			logger := slog.Default()
			logger = logger.With("path", *e.Path)

			logger.Debug("Iterate over directory contents")

			for rec, err := range it.Iterate(ctx, *e.Path) {

				if err != nil {
					logger.Warn("Failed to yield record", "error", err)
				}

				if !yield(rec, err) {
					return
				}
			}
		}
	}

}

func (it *GitHubAPIIterator) walkDirectoryContentsConcurrently(ctx context.Context, contents []*github.RepositoryContent) iter.Seq2[*iterate.Record, error] {

	return func(yield func(rec *iterate.Record, err error) bool) {

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		done_ch := make(chan bool)
		err_ch := make(chan error)
		rec_ch := make(chan *iterate.Record)

		for _, e := range contents {

			go func(e *github.RepositoryContent) {

				defer func() {
					done_ch <- true
				}()

				for rec, err := range it.Iterate(ctx, *e.Path) {

					select {
					case <-ctx.Done():
						break
					default:

						if err != nil {
							err_ch <- err
						} else {
							rec_ch <- rec
						}
					}
				}

			}(e)
		}

		remaining := len(contents)

		for remaining > 0 {
			select {
			case <-done_ch:
				remaining -= 1
			case err := <-err_ch:
				if !yield(nil, err) {
					return
				}
			case rec := <-rec_ch:
				if !yield(rec, nil) {
					return
				}
			}
		}

	}
}

func (it *GitHubAPIIterator) walkFileContents(ctx context.Context, contents *github.RepositoryContent) (*iterate.Record, error) {

	path := *contents.Path

	body, err := contents.GetContent()

	if err != nil {
		return nil, fmt.Errorf("Failed to read contents for %s, %w", path, err)
	}

	str_r := strings.NewReader(body)

	rsc, err := ioutil.NewReadSeekCloser(str_r)

	if err != nil {
		return nil, fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", path, err)
	}

	if it.filters != nil {

		ok, err := iterate.ApplyFilters(ctx, rsc, it.filters)

		if err != nil {

			rsc.Close()
			return nil, fmt.Errorf("Failed to apply filters for %s, %w", path, err)
		}

		if !ok {
			rsc.Close()
			return nil, nil
		}

	}

	return iterate.NewRecord(path, rsc), nil
}
