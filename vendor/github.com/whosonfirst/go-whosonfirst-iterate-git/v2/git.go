package git

import (
	"context"
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/filters"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func init() {
	ctx := context.Background()
	emitter.RegisterEmitter(ctx, "git", NewGitEmitter)
}

// GitEmitter implements the `Emitter` interface for crawling records in a Git repository.
type GitEmitter struct {
	emitter.Emitter
	// An optional path on disk where Git respositories will be cloned.
	target string
	// A boolean value indicating whether a Git repository (cloned to disk) should not be removed after processing.
	preserve bool
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// The branch of the Git repository to clone.
	branch string
}

// NewGitEmitter() returns a new `GitEmitter` instance configured by 'uri' in the form of:
//
//	git://{PATH}?{PARAMETERS}
//
// Where {PATH} is an optional path on disk where a repository will be clone to (default is to clone repository in memory) and {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
// * `?preserve=` A boolean value indicating whether a Git repository (cloned to disk) should not be removed after processing.
func NewGitEmitter(ctx context.Context, uri string) (emitter.Emitter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	em := &GitEmitter{
		target: u.Path,
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

	return em, nil
}

// WalkURI() walks (crawls) the Git repository identified by 'uri' and for each file (not excluded by any filters specified
// when `idx` was created) invokes 'index_cb'.
func (em *GitEmitter) WalkURI(ctx context.Context, index_cb emitter.EmitterCallbackFunc, uri string) error {

	var repo *gogit.Repository

	opts := &gogit.CloneOptions{
		URL: uri,
	}

	if em.branch != "" {
		br := plumbing.NewBranchReferenceName(em.branch)
		opts.ReferenceName = br
	}

	switch em.target {
	case "":

		r, err := gogit.Clone(memory.NewStorage(), nil, opts)

		if err != nil {
			return fmt.Errorf("Failed to clone repository, %w", err)
		}

		repo = r
	default:

		fname := filepath.Base(uri)
		path := filepath.Join(em.target, fname)

		r, err := gogit.PlainClone(path, false, opts)

		if err != nil {
			return fmt.Errorf("Failed to clone repository, %w", err)
		}

		if !em.preserve {
			defer os.RemoveAll(path)
		}

		repo = r
	}

	ref, err := repo.Head()

	if err != nil {
		return fmt.Errorf("Failed to derive head for repository, %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())

	if err != nil {
		return fmt.Errorf("Failed to derive object for ref hash, %w", err)
	}

	tree, err := commit.Tree()

	if err != nil {
		return fmt.Errorf("Failed to derive commit tree, %w", err)
	}

	err = tree.Files().ForEach(func(f *object.File) error {

		switch filepath.Ext(f.Name) {
		case ".geojson":
			// continue
		default:
			return nil
		}

		r, err := f.Reader()

		if err != nil {
			return fmt.Errorf("Failed to derive reader for %s, %w", f.Name, err)
		}

		defer r.Close()

		fh, err := ioutil.NewReadSeekCloser(r)

		if err != nil {
			return fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", f.Name, err)
		}

		defer fh.Close()
		
		if em.filters != nil {

			ok, err := em.filters.Apply(ctx, fh)

			if err != nil {
				return fmt.Errorf("Failed to apply query filters to %s, %w", f.Name, err)
			}

			if !ok {
				return nil
			}

			_, err = fh.Seek(0, 0)

			if err != nil {
				return fmt.Errorf("Failed to reset filehandle for %s, %w", f.Name, err)
			}
		}

		return index_cb(ctx, f.Name, fh)
	})

	if err != nil {
		return fmt.Errorf("Failed to iterate through tree, %w", err)
	}

	return nil
}
