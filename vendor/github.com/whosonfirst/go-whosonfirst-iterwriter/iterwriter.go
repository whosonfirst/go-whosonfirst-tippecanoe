package iterwriter

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

// Type IterwriterCallbackFunc defines a function to be invoked that will return a `emitter.EmitterCallbackFunc`
// callback function for use by `go-whosonfirst-iterate/v2/iterator` instances.
type IterwriterCallbackFunc func(writer.Writer, timings.Monitor) emitter.EmitterCallbackFunc

// DefaultIterwriterCallback returns a `emitter.EmitterCallbackFunc` callback function that will write each feature
// to a WOF-normalized path using 'wr'.
func DefaultIterwriterCallback(wr writer.Writer, monitor timings.Monitor) emitter.EmitterCallbackFunc {
	return standardIterwriterCallback(wr, monitor, false)
}

// ForgivingIterwriterCallback returns a `emitter.EmitterCallbackFunc` callback function that will write each feature
// to a WOF-normalized path using 'wr' but which will not return errors on write-failures (it will only log the errors
// to the default `slog.Logger` instance).
func ForgivingIterwriterCallback(wr writer.Writer, monitor timings.Monitor) emitter.EmitterCallbackFunc {
	return standardIterwriterCallback(wr, monitor, true)
}

func standardIterwriterCallback(wr writer.Writer, monitor timings.Monitor, forgiving bool) emitter.EmitterCallbackFunc {

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		logger := slog.Default()
		logger = logger.With("path", path)

		// See this? It's important. We are rewriting path to a normalized WOF relative path
		// That means this will only work with WOF documents

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			slog.Error("Failed to parse URI", "error", err)
			return fmt.Errorf("Unable to parse %s, %w", path, err)
		}

		logger = logger.With("id", id)

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			slog.Error("Failed to derive URI", "error", err)
			return fmt.Errorf("Unable to derive relative (WOF) path for %s, %w", path, err)
		}

		logger = logger.With("rel_path", rel_path)

		_, err = wr.Write(ctx, rel_path, r)

		if err != nil {

			slog.Error("Failed to write record", "error", err)

			if !forgiving {
				return fmt.Errorf("Failed to write record for %s, %w", rel_path, err)
			}
		}

		go monitor.Signal(ctx)
		return nil
	}

	return iter_cb
}

// IterateWithWriterAndCallback will process all files emitted by 'iterator_uri' and 'iterator_paths' passing each to 'iter_cb'
// (which it is assumed will eventually pass the file to 'wr').
func IterateWithWriterAndCallback(ctx context.Context, wr writer.Writer, iter_cb emitter.EmitterCallbackFunc, monitor timings.Monitor, iterator_uri string, iterator_paths ...string) error {

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		return fmt.Errorf("Failed to create new iterator, %w", err)
	}

	err = iter.IterateURIs(ctx, iterator_paths...)

	if err != nil {
		return fmt.Errorf("Failed to iterate paths, %w", err)
	}

	err = wr.Close(ctx)

	if err != nil {
		return fmt.Errorf("Failed to close writer, %w", err)
	}

	return nil
}
