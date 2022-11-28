package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-organization"
	_ "github.com/whosonfirst/go-writer-featurecollection/v3"
	_ "github.com/whosonfirst/go-writer-jsonl/v3"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-tippecanoe/app/features"
	"log"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := features.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run iterwriter, %v", err)
	}
}
