package tippecanoe

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-iterwriter/v4"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

type IterwriterCallbackFuncBuilderOptions struct {
	RequirePolygon      bool
	AsSPR               bool
	IncludeAltFiles     bool
	AppendSPRProperties []string
	Forgiving           bool
}

func IterwriterCallbackFuncBuilder(opts *IterwriterCallbackFuncBuilderOptions) iterwriter.IterwriterCallback {

	// wr, logger, monitor are created in whosonfirst/go-whosonfirst-iterwriter/app/iterwriter

	fn := func(ctx context.Context, rec *iterate.Record, wr writer.Writer) error {

		logger := slog.Default()
		logger = logger.With("rec.Path", rec.Path)

		id, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			return fmt.Errorf("Unable to parse %s, %w", rec.Path, err)
		}

		logger = logger.With("id", id)

		if uri_args.IsAlternate && !opts.IncludeAltFiles {
			logger.Debug("Is alternate geometry, skipping")
			return nil
		}

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			return fmt.Errorf("Unable to derive relative (WOF) path for %s, %w", rec.Path, err)
		}

		logger = logger.With("rel_path", rel_path)

		var wr_body io.ReadSeeker
		wr_body = rec.Body

		if opts.RequirePolygon || opts.AsSPR {

			body, err := io.ReadAll(rec.Body)

			if err != nil {
				logger.Error("Failed to read body", "error", err)

				if opts.Forgiving {
					return nil
				}

				return fmt.Errorf("Failed to read body for %s, %w", rec.Path, err)
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

					logger.Error("Failed to derive SPR", "error", err)

					if opts.Forgiving {
						return nil
					}

					return fmt.Errorf("Failed to create SPR for %s, %w", rec.Path, err)
				}

				// ideally use go-whosonfirst-spatial.PropertiesResponseResultsWithStandardPlacesResults here
				// but not sure what the what is yet

				old_props := gjson.GetBytes(body, "properties")

				body, err = sjson.SetBytes(body, "properties", s)

				if err != nil {
					logger.Error("Failed to update properties with SPR", "error", err)

					if opts.Forgiving {
						return nil
					}

					return fmt.Errorf("Failed to update properties for %s, %w", rec.Path, err)
				}

				if len(opts.AppendSPRProperties) > 0 {

					for _, path := range opts.AppendSPRProperties {

						// Because we are deriving this from old_props and not body
						rel_path := strings.Replace(path, "properties.", "", 1)

						p_rsp := old_props.Get(rel_path)

						if p_rsp.Exists() {

							abs_path := fmt.Sprintf("properties.%s", rel_path)

							body, err = sjson.SetBytes(body, abs_path, p_rsp.Value())

							if err != nil {

								logger.Error("Failed to append property to SPR", "property", abs_path, "error", err)

								if opts.Forgiving {
									return nil
								}

								return fmt.Errorf("Failed to assign %s to properties, %w", abs_path, err)
							}
						}
					}

				}
			}

			br := bytes.NewReader(body)
			wr_body = br
		}

		_, err = wr.Write(ctx, rel_path, wr_body)

		if err != nil {

			logger.Error("Failed to write document", "error", err)

			if opts.Forgiving {
				return nil
			}

			return fmt.Errorf("Failed to write %s, %v", rec.Path, err)
		}

		return nil
	}

	return fn
}
