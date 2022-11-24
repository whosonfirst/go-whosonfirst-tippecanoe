package iterwriter

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-timings"
	"github.com/sfomuseum/runtimevar"
	"github.com/whosonfirst/go-whosonfirst-iterwriter"
	"github.com/whosonfirst/go-writer/v3"
	"log"
	"os"
	"strings"
	"time"
)

func Run(ctx context.Context, logger *log.Logger) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WOF")

	if err != nil {
		return fmt.Errorf("Failed to assign flags from environment variables, %w", err)
	}

	iterator_paths := fs.Args()

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

	mw, err := writer.NewMultiWriter(ctx, writers...)

	if err != nil {
		return fmt.Errorf("Failed to create multi writer, %w", err)
	}

	monitor, err := timings.NewMonitor(ctx, monitor_uri)

	if err != nil {
		return fmt.Errorf("Failed to create monitor, %w", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	err = iterwriter.IterateWithWriter(ctx, mw, monitor, iterator_uri, iterator_paths...)

	if err != nil {
		return fmt.Errorf("Failed to iterate with writer, %w", err)
	}

	return nil
}
