package iterwriter

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/v4"
	"github.com/whosonfirst/go-writer/v3"
)

type RunOptions struct {
	CallbackFunc  iterwriter.IterwriterCallback
	Writer        writer.Writer
	IteratorURI   string
	IteratorPaths []string
	MonitorURI    string
	MonitorWriter io.Writer
	Verbose       bool
}

func DefaultOptionsFromFlagSet(fs *flag.FlagSet, parsed bool) (*RunOptions, error) {

	if !parsed {

		flagset.Parse(fs)

		err := flagset.SetFlagsFromEnvVars(fs, "WOF")

		if err != nil {
			return nil, fmt.Errorf("Failed to assign flags from environment variables, %w", err)
		}
	}

	cb_func := iterwriter.DefaultIterwriterCallback(forgiving)

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
