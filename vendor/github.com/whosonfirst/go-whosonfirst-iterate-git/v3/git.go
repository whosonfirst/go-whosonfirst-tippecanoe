package git

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

func init() {
	ctx := context.Background()
	iterate.RegisterIterator(ctx, "git", NewGitIterator)
}

// GitIterator implements the `Iterator` interface for crawling records in a Git repository.
type GitIterator struct {
	iterate.Iterator
	// An optional path on disk where Git respositories will be cloned.
	target string
	// A boolean value indicating whether a Git repository (cloned to disk) should not be removed after processing.
	preserve bool
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// The branch of the Git repository to clone.
	branch string
	// Limit fetching to the specified number of commits.
	depth int
	// The count of documents that have been processed so far.
	seen int64
	// Boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewGitIterator() returns a new `GitIterator` instance configured by 'uri' in the form of:
//
//	git://{PATH}?{PARAMETERS}
//
// Where {PATH} is an optional path on disk where a repository will be clone to (default is to clone repository in memory) and {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
// * `?preserve=` A boolean value indicating whether a Git repository (cloned to disk) should not be removed after processing.
// * `?depth=` An integer value indicating the number of commits to fetch. Default is 1.
func NewGitIterator(ctx context.Context, uri string) (iterate.Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	em := &GitIterator{
		target:    u.Path,
		depth:     1,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	q := u.Query()

	f, err := filters.NewQueryFiltersFromQuery(ctx, q)

	if err != nil {
		return nil, fmt.Errorf("Failed to create query filters, %w", err)
	}

	em.filters = f

	str_preserve := q.Get("preserve")

	if str_preserve != "" {

		preserve, err := strconv.ParseBool(str_preserve)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse 'preserve' parameter, %w", err)
		}

		em.preserve = preserve
	}

	branch := q.Get("branch")

	if branch != "" {
		em.branch = branch
	}

	if q.Has("depth") {

		v, err := strconv.Atoi(q.Get("depth"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '?depth=' parameter, %w", err)
		}

		em.depth = v
	}

	return em, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *GitIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*iterate.Record, error] {

	return func(yield func(rec *iterate.Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			logger := slog.Default()
			logger = logger.With("uri", uri)
			logger = logger.With("branch", it.branch)

			var repo *gogit.Repository

			clone_opts := &gogit.CloneOptions{
				URL:   uri,
				Depth: it.depth,
			}

			if it.branch != "" {
				br := plumbing.NewBranchReferenceName(it.branch)
				clone_opts.ReferenceName = br
			}

			logger = logger.With("target", it.target)

			t1 := time.Now()

			switch it.target {
			case "":

				logger.Debug("Clone in to memory")

				r, err := gogit.Clone(memory.NewStorage(), nil, clone_opts)

				if err != nil {
					logger.Error("Failed to clone repo", "error", err)

					if !yield(nil, err) {
						return
					}

					continue
				}

				repo = r
			default:

				fname := filepath.Base(uri)
				path := filepath.Join(it.target, fname)

				logger.Debug("Clone to path", "path", path)

				r, err := gogit.PlainClone(path, false, clone_opts)

				if err != nil {
					logger.Error("Failed to clone repo", "error", err)

					if !yield(nil, err) {
						return
					}

					continue
				}

				if !it.preserve {
					defer os.RemoveAll(path)
				}

				repo = r
			}

			logger.Debug("Time to clone repo", "time", time.Since(t1))

			ref, err := repo.Head()

			if err != nil {
				logger.Error("Failed to derive HEAD", "error", err)

				if !yield(nil, err) {
					return
				}

				continue
			}

			logger = logger.With("ref", ref.Hash())

			commit, err := repo.CommitObject(ref.Hash())

			if err != nil {
				logger.Error("Failed to derive commit object", "error", err)

				if !yield(nil, err) {
					return
				}

				continue
			}

			tree, err := commit.Tree()

			if err != nil {
				logger.Error("Failed to derive commit tree", "error", err)

				if !yield(nil, err) {
					return
				}

				continue
			}

			err = tree.Files().ForEach(func(f *object.File) error {

				logger := slog.Default()
				logger = logger.With("uri", uri)
				logger = logger.With("path", f.Name)
				logger = logger.With("branch", it.branch)
				logger = logger.With("ref", ref.Hash())

				switch filepath.Ext(f.Name) {
				case ".geojson":
					// continue
				default:
					// logger.Debug("Not a .geojson file, skipping.")
					return nil
				}

				r, err := f.Reader()

				if err != nil {
					logger.Error("Failed to derive reader", "error", err)
					return fmt.Errorf("Failed to derive reader for %s, %w", f.Name, err)
				}

				rsc, err := ioutil.NewReadSeekCloser(r)

				if err != nil {
					r.Close()
					logger.Error("Failed to create ReadSeekCloser", "error", err)
					return fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", f.Name, err)
				}

				if it.filters != nil {

					ok, err := it.filters.Apply(ctx, rsc)

					if err != nil {
						rsc.Close()
						logger.Error("Failed to apply filters", "error", err)
						return fmt.Errorf("Failed to apply query filters to %s, %w", f.Name, err)
					}

					if !ok {
						// logger.Debug("Skipping because filters not true")
						rsc.Close()
						return nil
					}

					_, err = rsc.Seek(0, 0)

					if err != nil {
						rsc.Close()
						logger.Error("Failed to rewind filehandler", "error", err)
						return fmt.Errorf("Failed to reset filehandle for %s, %w", f.Name, err)
					}
				}

				rec := iterate.NewRecord(f.Name, rsc)

				yield(rec, nil)
				return nil
			})

			if err != nil {
				logger.Error("Failed to iterate tree", "error", err)
				yield(nil, err)
				return
			}

		}
	}
}

// Seen() returns the total number of records processed so far.
func (it *GitIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *GitIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *GitIterator) Close() error {
	return nil
}
