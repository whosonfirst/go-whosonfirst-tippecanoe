package iterwriter

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-timings"
	"github.com/sfomuseum/runtimevar"
	"github.com/whosonfirst/go-whosonfirst-iterwriter"
	"github.com/whosonfirst/go-writer/v3"
)

type RunOptions struct {
	CallbackFunc  iterwriter.IterwriterCallbackFunc
	Writer        writer.Writer
	IteratorURI   string
	IteratorPaths []string
	MonitorURI    string
	MonitorWriter io.Writer
	Verbose       bool
}

func Run(ctx context.Context, logger *slog.Logger) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *slog.Logger) error {

	opts, err := DefaultOptionsFromFlagSet(fs, false)

	if err != nil {
		return fmt.Errorf("Failed to derive options from flags, %w", err)
	}

	return RunWithOptions(ctx, opts, logger)
}

func DefaultOptionsFromFlagSet(fs *flag.FlagSet, parsed bool) (*RunOptions, error) {

	if !parsed {

		flagset.Parse(fs)

		err := flagset.SetFlagsFromEnvVars(fs, "WOF")

		if err != nil {
			return nil, fmt.Errorf("Failed to assign flags from environment variables, %w", err)
		}
	}

	cb_func := iterwriter.DefaultIterwriterCallback

	if forgiving {
		cb_func = iterwriter.ForgivingIterwriterCallback
	}

	iterator_paths := fs.Args()

	opts := &RunOptions{
		CallbackFunc:  cb_func,
		IteratorURI:   iterator_uri,
		IteratorPaths: iterator_paths,
		MonitorURI:    monitor_uri,
		MonitorWriter: os.Stderr,
		Verbose:       verbose,
	}

	return opts, nil
}

func RunWithOptions(ctx context.Context, opts *RunOptions, logger *slog.Logger) error {

	slog.SetDefault(logger)

	if opts.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		logger.Debug("Verbose (debug) logging enabled")
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

	iter_cb := opts.CallbackFunc(mw, monitor)

	err = iterwriter.IterateWithWriterAndCallback(ctx, mw, iter_cb, monitor, opts.IteratorURI, opts.IteratorPaths...)

	if err != nil {
		return fmt.Errorf("Failed to iterate with writer, %w", err)
	}

	return nil
}
