package pgxgeos

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

// Register registers codecs for [github.com/twpayne/go-geos] types on conn.
func Register(ctx context.Context, conn *pgx.Conn, geosContext *geos.Context) error {
	return errors.Join(
		registerBox2D(ctx, conn),
		registerGeom(ctx, conn, geosContext),
	)
}
