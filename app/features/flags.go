package features

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/application/iterwriter"
)

var as_spr bool
var require_polygons bool
var include_alt_files bool

func DefaultFlagSet() *flag.FlagSet {

	fs := iterwriter.DefaultFlagSet()

	fs.BoolVar(&as_spr, "as-spr", true, "Replace Feature properties with Who's On First Standard Places Result (SPR) derived from that feature.")
	fs.BoolVar(&require_polygons, "require-polygons", false, "Require that geometry type be 'Polygon' or 'MultiPolygon' to be included in output.")
	fs.BoolVar(&include_alt_files, "include-alt-files", false, "Include alternate geometry files in output.")

	return fs
}
