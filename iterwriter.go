package tippecanoe

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
	"io"
	"log"
)

func IterwriterCallbackFunc(wr writer.Writer, logger *log.Logger, monitor timings.Monitor) emitter.EmitterCallbackFunc {

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		// See this? It's important. We are rewriting path to a normalized WOF relative path
		// That means this will only work with WOF documents

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Unable to parse %s, %w", path, err)
		}

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			return fmt.Errorf("Unable to derive relative (WOF) path for %s, %w", path, err)
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read body for %s, %w", path, err)
		}

		geom_rsp := gjson.GetBytes(body, "geometry.type")

		switch geom_rsp.String() {
		case "Polygon", "MultiPolygon":
			// pass
		default:
			return nil
		}

		var s spr.StandardPlacesResult

		if uri_args.IsAlternate {

			s, err = spr.WhosOnFirstAltSPR(body)
		} else {

			s, err = spr.WhosOnFirstSPR(body)

		}

		if err != nil {
			return fmt.Errorf("Failed to create SPR for %s, %w", path, err)
		}

		body, err = sjson.SetBytes(body, "properties", s)

		if err != nil {
			return fmt.Errorf("Failed to update properties for %s, %w", path, err)
		}

		br := bytes.NewReader(body)

		_, err = wr.Write(ctx, rel_path, br)

		if err != nil {
			return fmt.Errorf("Failed to write %s, %v", path, err)
		}

		go monitor.Signal(ctx)
		return nil
	}

	return iter_cb
}
