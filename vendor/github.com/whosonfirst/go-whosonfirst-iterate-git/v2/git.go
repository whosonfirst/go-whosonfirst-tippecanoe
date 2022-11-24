package git

import (
	"context"
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
)

func init() {
	ctx := context.Background()
	emitter.RegisterEmitter(ctx, "git", NewGitEmitter)
}

type GitEmitter struct {
	emitter.Emitter
	target   string
	preserve bool
	filters  filters.Filters
	branch   string
}

func NewGitEmitter(ctx context.Context, uri string) (emitter.Emitter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	em := &GitEmitter{
		target: u.Path,
	}

	q := u.Query()

	f, err := filters.NewQueryFiltersFromQuery(ctx, q)

	if err != nil {
		return nil, err
	}

	em.filters = f

	if q.Get("preserve") == "1" {
		em.preserve = true
	}

	branch := q.Get("branch")

	if branch != "" {
		em.branch = branch
	}

	return em, nil
}

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
			return err
		}

		repo = r
	default:

		fname := filepath.Base(uri)
		path := filepath.Join(em.target, fname)

		r, err := gogit.PlainClone(path, false, opts)

		if err != nil {
			return err
		}

		if !em.preserve {
			defer os.RemoveAll(path)
		}

		repo = r
	}

	ref, err := repo.Head()

	if err != nil {
		return err
	}

	commit, err := repo.CommitObject(ref.Hash())

	if err != nil {
		return err
	}

	tree, err := commit.Tree()

	if err != nil {
		return err
	}

	tree.Files().ForEach(func(f *object.File) error {

		switch filepath.Ext(f.Name) {
		case ".geojson":
			// continue
		default:
			return nil
		}

		r, err := f.Reader()

		if err != nil {
			return err
		}

		defer r.Close()

		fh, err := ioutil.NewReadSeekCloser(r)

		if err != nil {
			return err
		}

		if em.filters != nil {

			ok, err := em.filters.Apply(ctx, fh)

			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			_, err = fh.Seek(0, 0)

			if err != nil {
				return err
			}
		}

		return index_cb(ctx, f.Name, fh)
	})

	return nil
}
