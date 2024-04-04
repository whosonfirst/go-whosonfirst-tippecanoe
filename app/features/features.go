package features

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"	
	"github.com/whosonfirst/go-whosonfirst-iterwriter/app/iterwriter"
	"github.com/whosonfirst/go-whosonfirst-tippecanoe"	
)

func Run(ctx context.Context, logger *slog.Logger) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *slog.Logger) error {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WOF")
	
	if err != nil {
		return fmt.Errorf("Failed to assign flags from environment variables, %w", err)
	}
	
	opts, err := iterwriter.DefaultOptionsFromFlagSet(fs, true)

	if err != nil {
		return fmt.Errorf("Failed to derive default options from flags, %w", err)
	}

	forgiving, err := lookup.BoolVar(fs, "forgiving")

	if err != nil {
		return fmt.Errorf("Failed to derive forgiving flag, %w", err)
	}
	
	cb_opts := &tippecanoe.IterwriterCallbackFuncBuilderOptions{
		AsSPR:               as_spr,
		RequirePolygon:      require_polygons,
		IncludeAltFiles:     include_alt_files,
		AppendSPRProperties: spr_properties,
		Forgiving: forgiving,
	}

	cb_func := tippecanoe.IterwriterCallbackFuncBuilder(cb_opts)

	opts.CallbackFunc = cb_func

	err = iterwriter.RunWithOptions(ctx, opts, logger)

	if err != nil {
		return fmt.Errorf("Failed to run iterwriter, %v", err)
	}

	return nil
}
