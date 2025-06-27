package iterwriter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

type IterwriterCallback func(context.Context, *iterate.Record, writer.Writer) error

func DefaultIterwriterCallback(forgiving bool) IterwriterCallback {

	return func(ctx context.Context, rec *iterate.Record, wr writer.Writer) error {

		logger := slog.Default()
		logger = logger.With("path", rec.Path)

		// See this? It's important. We are rewriting path to a normalized WOF relative path
		// That means this will only work with WOF documents

		id, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			slog.Error("Failed to parse URI", "error", err)
			return fmt.Errorf("Unable to parse %s, %w", rec.Path, err)
		}

		logger = logger.With("id", id)

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			slog.Error("Failed to derive URI", "error", err)
			return fmt.Errorf("Unable to derive relative (WOF) path for %s, %w", rec.Path, err)
		}

		logger = logger.With("rel_path", rel_path)

		_, err = wr.Write(ctx, rel_path, rec.Body)

		if err != nil {

			slog.Error("Failed to write record", "error", err)

			if !forgiving {
				return fmt.Errorf("Failed to write record for %s, %w", rel_path, err)
			}
		}

		return nil
	}
}
