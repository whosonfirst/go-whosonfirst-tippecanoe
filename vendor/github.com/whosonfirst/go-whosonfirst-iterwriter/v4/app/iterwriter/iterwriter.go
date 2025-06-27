package iterwriter

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sfomuseum/go-timings"
	"github.com/sfomuseum/runtimevar"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-writer/v3"
)

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	opts, err := DefaultOptionsFromFlagSet(fs, false)

	if err != nil {
		return fmt.Errorf("Failed to derive options from flags, %w", err)
	}

	return RunWithOptions(ctx, opts)
}

func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	if opts.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose (debug) logging enabled")
	}

	var mw writer.Writer

	if opts.Writer != nil {
		mw = opts.Writer
	} else {

		writers := make([]writer.Writer, len(writer_uris))

		wr_ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		for idx, runtimevar_uri := range writer_uris {

			wr_uri, err := runtimevar.StringVar(wr_ctx, runtimevar_uri)

			if err != nil {
				return fmt.Errorf("Failed to derive writer URI for %s, %w", runtimevar_uri, err)
			}

			wr_uri = strings.TrimSpace(wr_uri)

			wr, err := writer.NewWriter(ctx, wr_uri)

			if err != nil {

				return fmt.Errorf("Failed to create new writer for %s, %w", runtimevar_uri, err)
			}

			writers[idx] = wr
		}

		_mw, err := writer.NewMultiWriter(ctx, writers...)

		if err != nil {
			return fmt.Errorf("Failed to create multi writer, %w", err)
		}

		mw = _mw
	}

	monitor, err := timings.NewMonitor(ctx, opts.MonitorURI)

	if err != nil {
		return fmt.Errorf("Failed to create monitor, %w", err)
	}

	monitor.Start(ctx, opts.MonitorWriter)
	defer monitor.Stop(ctx)

	iter, err := iterate.NewIterator(ctx, opts.IteratorURI)

	if err != nil {
		return fmt.Errorf("Failed to create new iterator, %w", err)
	}

	for rec, err := range iter.Iterate(ctx, opts.IteratorPaths...) {

		if err != nil {
			return err
		}

		defer rec.Body.Close()

		err = opts.CallbackFunc(ctx, rec, mw)

		if err != nil {
			return err
		}

		monitor.Signal(ctx)
	}

	err = mw.Close(ctx)

	if err != nil {
		return fmt.Errorf("Failed to close writer, %w", err)
	}

	err = iter.Close()

	if err != nil {
		return fmt.Errorf("Failed to close iterator, %w", err)
	}

	return nil
}
