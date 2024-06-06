package pgxgeos

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

// Register registers codecs for [github.com/twpayne/go-geos] types on conn.
func Register(ctx context.Context, conn *pgx.Conn, geosContext *geos.Context) error {
	if err := registerGeom(ctx, conn, geosContext); err != nil {
		return err
	}

	return nil
}
