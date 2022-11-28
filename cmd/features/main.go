package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-organization"
	_ "github.com/whosonfirst/go-writer-featurecollection/v3"
	_ "github.com/whosonfirst/go-writer-jsonl/v3"
)

import (
	"context"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/application/iterwriter"
	"github.com/whosonfirst/go-whosonfirst-tippecanoe"
	"log"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	fs := iterwriter.DefaultFlagSet()

	as_spr := fs.Bool("as-spr", true, "Replace Feature properties with Who's On First Standard Places Result (SPR) derived from that feature.")
	require_polygons := fs.Bool("require-polygons", false, "Require that geometry type be 'Polygon' or 'MultiPolygon' to be included in output.")
	include_alt_files := fs.Bool("include-alt-files", false, "Include alternate geometry files in output.")
	
	flagset.Parse(fs)

	cb_opts := &tippecanoe.IterwriterCallbackFuncBuilderOptions{
		AsSPR:          *as_spr,
		RequirePolygon: *require_polygons,
		IncludeAltFiles: *include_alt_files,
	}

	cb := tippecanoe.IterwriterCallbackFuncBuilder(cb_opts)

	opts := &iterwriter.RunOptions{
		Logger:       logger,
		FlagSet:      fs,
		CallbackFunc: cb,
	}

	err := iterwriter.RunWithOptions(ctx, opts)

	if err != nil {
		logger.Fatalf("Failed to run iterwriter, %v", err)
	}
}
