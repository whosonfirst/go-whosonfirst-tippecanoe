package iterwriter

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

var writer_uris multi.MultiCSVString
var iterator_uri string
var monitor_uri string
var forgiving bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("es")

	fs.Var(&writer_uris, "writer-uri", "One or more valid whosonfirst/go-writer/v2 URIs, each encoded as a gocloud.dev/runtimevar URI.")
	fs.StringVar(&iterator_uri, "iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v3 URI.")
	fs.StringVar(&monitor_uri, "monitor-uri", "counter://PT60S", "A valid sfomuseum/go-timings URI.")
	fs.BoolVar(&forgiving, "forgiving", false, "Be \"forgiving\" of failed writes, logging the issue(s) but not triggering errors")
	return fs
}
