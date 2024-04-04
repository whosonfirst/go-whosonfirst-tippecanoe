package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-organization"
	_ "github.com/whosonfirst/go-writer-featurecollection/v3"
	_ "github.com/whosonfirst/go-writer-jsonl/v3"
)

import (
	"context"
	"log/slog"
	"os"
	
	"github.com/whosonfirst/go-whosonfirst-tippecanoe/app/features"	
)

func main() {

	ctx := context.Background()
	logger := slog.Default()

	err := features.Run(ctx, logger)

	if err != nil {
		logger.Error("Failed to run iterwriter", "error", err)
		os.Exit(1)
	}
}
