package organizations

import (
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/whosonfirst/go-whosonfirst-github/util"
	"log"
	"strings"
	"time"
)

type ListOptions struct {
	Prefix      []string
	Exclude     []string
	Forked      bool
	NotForked   bool
	AccessToken string
	PushedSince *time.Time
	Debug       bool
}

type CreateOptions struct {
	AccessToken string
	Name        string
	Description string
	Private     bool
}

func NewDefaultListOptions() *ListOptions {

	opts := ListOptions{
		Prefix:      []string{},
		Exclude:     []string{},
		Forked:      false,
		NotForked:   false,
		AccessToken: "",
		PushedSince: nil,
		Debug:       false,
	}

	return &opts
}

// CreateRepo is a helper method for creating a new
func CreateRepo(org_name string, opts *CreateOptions) error {

	// https://docs.github.com/en/rest/reference/repos#create-an-organization-repository
	// https://github.com/google/go-github/blob/v17.0.0/example/newrepo/main.go
	// https://github.com/google/go-github/blob/v17.0.0/github/repos.go#L262

	client, ctx, err := util.NewClientAndContext(opts.AccessToken)

	if err != nil {
		return fmt.Errorf("Failed to create new client, %w", err)
	}

	r := &github.Repository{
		Name:        &opts.Name,
		Private:     &opts.Private,
		Description: &opts.Description,
	}

	_, _, err = client.Repositories.Create(ctx, org_name, r)

	if err != nil {
		return fmt.Errorf("Failed to create repository, %w", err)
	}

	return nil
}

func ListRepos(org string, opts *ListOptions) ([]string, error) {

	repos := make([]string, 0)

	cb := func(r *github.Repository) error {
		repos = append(repos, *r.Name)
		return nil
	}

	err := ListReposWithCallback(org, opts, cb)

	return repos, err
}

func ListReposWithCallback(org string, opts *ListOptions, cb func(repo *github.Repository) error) error {

	client, ctx, err := util.NewClientAndContext(opts.AccessToken)

	if err != nil {
		return err
	}

	gh_opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		possible, resp, err := client.Repositories.ListByOrg(ctx, org, gh_opts)

		if err != nil {
			return err
		}

		for _, r := range possible {

			if len(opts.Prefix) > 0 {

				has_prefix := false

				for _, prefix := range opts.Prefix {
					if strings.HasPrefix(*r.Name, prefix) {
						has_prefix = true
						break
					}
				}

				if !has_prefix {
					continue
				}
			}

			if len(opts.Exclude) > 0 {

				is_excluded := false

				for _, prefix := range opts.Exclude {

					if strings.HasPrefix(*r.Name, prefix) {
						is_excluded = true
						break
					}
				}

				if is_excluded {
					continue
				}
			}

			if opts.Forked && !*r.Fork {
				continue
			}

			if opts.NotForked && *r.Fork {
				continue
			}

			if opts.PushedSince != nil {

				if opts.Debug {
					log.Printf("SINCE %s pushed at %v (%v) : %t\n", *r.Name, r.PushedAt, *opts.PushedSince, r.PushedAt.Before(*opts.PushedSince))
				}

				if r.PushedAt.Before(*opts.PushedSince) {
					continue
				}
			}

			err := cb(r)

			if err != nil {
				return fmt.Errorf("Failed to invoke callback for '%s', %w", *r.Name, err)
			}

		}

		if resp.NextPage == 0 {
			break
		}

		gh_opts.ListOptions.Page = resp.NextPage
	}

	return nil
}
