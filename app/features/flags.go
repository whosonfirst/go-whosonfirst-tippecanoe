package features

import (
	"flag"
	"github.com/sfomuseum/go-flags/multi"

	"github.com/whosonfirst/go-whosonfirst-iterwriter/v4/app/iterwriter"
)

var as_spr bool
var require_polygons bool
var include_alt_files bool

var spr_properties multi.MultiCSVString

func DefaultFlagSet() *flag.FlagSet {

	fs := iterwriter.DefaultFlagSet()

	fs.BoolVar(&as_spr, "as-spr", true, "Replace Feature properties with Who's On First Standard Places Result (SPR) derived from that feature.")
	fs.BoolVar(&require_polygons, "require-polygons", false, "Require that geometry type be 'Polygon' or 'MultiPolygon' to be included in output.")
	fs.BoolVar(&include_alt_files, "include-alt-files", false, "Include alternate geometry files in output.")

	fs.Var(&spr_properties, "spr-append-property", "Zero or more properties in a given feature to append to SPR output")
	return fs
}
