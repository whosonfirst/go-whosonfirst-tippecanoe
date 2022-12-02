package features

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/application/iterwriter"
	"github.com/whosonfirst/go-whosonfirst-tippecanoe"
	"log"
)

func Run(ctx context.Context, logger *log.Logger) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	flagset.Parse(fs)

	cb_opts := &tippecanoe.IterwriterCallbackFuncBuilderOptions{
		AsSPR:               as_spr,
		RequirePolygon:      require_polygons,
		IncludeAltFiles:     include_alt_files,
		AppendSPRProperties: spr_properties,
	}

	cb := tippecanoe.IterwriterCallbackFuncBuilder(cb_opts)

	opts := &iterwriter.RunOptions{
		Logger:       logger,
		FlagSet:      fs,
		CallbackFunc: cb,
	}

	err := iterwriter.RunWithOptions(ctx, opts)

	if err != nil {
		return fmt.Errorf("Failed to run iterwriter, %v", err)
	}

	return nil
}
