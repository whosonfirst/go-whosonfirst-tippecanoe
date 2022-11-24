package main

import (
	_ "github.com/whosonfirst/go-writer-featurecollection/v3"
)

import (
	"github.com/whosonfirst/go-whosonfirst-iterwriter/application/iterwriter"
	"context"
	"log"
)

func main(){

	ctx := context.Background()
	logger := log.Default()

	err := iterwriter.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run iterwriter, %v", err)
	}
}
