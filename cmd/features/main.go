package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-iterate-organization/v2"
	_ "github.com/whosonfirst/go-writer-featurecollection/v3"
	_ "github.com/whosonfirst/go-writer-jsonl/v3"

	"github.com/whosonfirst/go-whosonfirst-tippecanoe/app/features"
)

func main() {

	ctx := context.Background()
	err := features.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to derive features, %v", err)
	}
}
