package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-organization"
	_ "github.com/whosonfirst/go-writer-featurecollection/v3"
	_ "github.com/whosonfirst/go-writer-jsonl/v3"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/application/iterwriter"
	"github.com/whosonfirst/go-whosonfirst-tippecanoe"
	"log"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	fs := iterwriter.DefaultFlagSet()

	opts := &iterwriter.RunOptions{
		Logger:       logger,
		FlagSet:      fs,
		CallbackFunc: tippecanoe.IterwriterCallbackFunc,
	}

	err := iterwriter.RunWithOptions(ctx, opts)

	if err != nil {
		logger.Fatalf("Failed to run iterwriter, %v", err)
	}
}
