package features

import (
	"context"
	"flag"
	"fmt"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/v3/app/iterwriter"
	"github.com/whosonfirst/go-whosonfirst-tippecanoe"
)

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

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
		Forgiving:           forgiving,
	}

	cb_func := tippecanoe.IterwriterCallbackFuncBuilder(cb_opts)

	opts.CallbackFunc = cb_func

	err = iterwriter.RunWithOptions(ctx, opts)

	if err != nil {
		return fmt.Errorf("Failed to run iterwriter, %v", err)
	}

	return nil
}
