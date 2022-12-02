package tippecanoe

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterwriter"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
	"io"
	"log"
	"strings"
)

type IterwriterCallbackFuncBuilderOptions struct {
	RequirePolygon   bool
	AsSPR            bool
	IncludeAltFiles  bool
	AppendPropertiesToSPR []string
}

func IterwriterCallbackFuncBuilder(opts *IterwriterCallbackFuncBuilderOptions) iterwriter.IterwriterCallbackFunc {

	// wr, logger, monitor are created in whosonfirst/go-whosonfirst-iterwriter/app/iterwriter

	iterwriter_cb := func(wr writer.Writer, logger *log.Logger, monitor timings.Monitor) emitter.EmitterCallbackFunc {

		iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

			id, uri_args, err := uri.ParseURI(path)

			if err != nil {
				return fmt.Errorf("Unable to parse %s, %w", path, err)
			}

			if uri_args.IsAlternate && !opts.IncludeAltFiles {
				return nil
			}

			rel_path, err := uri.Id2RelPath(id, uri_args)

			if err != nil {
				return fmt.Errorf("Unable to derive relative (WOF) path for %s, %w", path, err)
			}

			if opts.RequirePolygon || opts.AsSPR {

				body, err := io.ReadAll(r)

				if err != nil {
					return fmt.Errorf("Failed to read body for %s, %w", path, err)
				}

				if opts.RequirePolygon {

					geom_rsp := gjson.GetBytes(body, "geometry.type")

					switch geom_rsp.String() {
					case "Polygon", "MultiPolygon":
						// pass
					default:
						return nil
					}
				}

				if opts.AsSPR && !uri_args.IsAlternate {

					s, err := spr.WhosOnFirstSPR(body)

					if err != nil {
						return fmt.Errorf("Failed to create SPR for %s, %w", path, err)
					}

					// ideally use go-whosonfirst-spatial.PropertiesResponseResultsWithStandardPlacesResults here
					// but not sure what the what is yet

					old_props := gjson.GetBytes(body, "properties")

					body, err = sjson.SetBytes(body, "properties", s)

					if err != nil {
						return fmt.Errorf("Failed to update properties for %s, %w", path, err)
					}

					if len(opts.AppendPropertiesToSPR) > 0 {

						for _, path := range opts.AppendPropertiesToSPR {

							rel_path := strings.Replace(path, "properties.", "", 1)

							p_rsp := old_props.Get(rel_path)

							if p_rsp.Exists() {

								abs_path := fmt.Sprintf("properties.%s", rel_path)

								body, err = sjson.SetBytes(body, abs_path, p_rsp.Value())

								if err != nil {
									return fmt.Errorf("Failed to assign %s to properties, %w", abs_path, err)
								}
							}
						}

					}
				}

				br := bytes.NewReader(body)
				r = br
			}

			_, err = wr.Write(ctx, rel_path, r)

			if err != nil {
				return fmt.Errorf("Failed to write %s, %v", path, err)
			}

			go monitor.Signal(ctx)
			return nil
		}

		return iter_cb
	}

	return iterwriter_cb
}
